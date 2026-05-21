package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"github.com/urfave/cli/v3"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	cmd := &cli.Command{
		Name: "cloudbuild-cancel-in-progress",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "project-id",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "location",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "build-id",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "filter",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return cancelInProgressBuilds(
				ctx,
				c.String("project-id"),
				c.String("location"),
				c.String("build-id"),
				c.String("filter"),
			)
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func cancelInProgressBuilds(
	ctx context.Context,
	projectID,
	location,
	buildID,
	filter string,
) (err error) {
	client, err := cloudbuild.NewClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if e := client.Close(); e != nil && err == nil {
			err = e
		}
	}()

	var createTimeOfCurrentBuild string
	if currentBuild, err := client.GetBuild(ctx, &cloudbuildpb.GetBuildRequest{
		Name: "projects/" + projectID + "/locations/" + location + "/builds/" + buildID,
	}); err != nil {
		return err
	} else {
		createTimeOfCurrentBuild = currentBuild.GetCreateTime().AsTime().Format(time.RFC3339Nano)
	}
	filter = fmt.Sprintf(`(status="QUEUED" OR status="WORKING" OR status="PENDING") AND create_time<=%q AND build_id!=%q AND (%s)`, createTimeOfCurrentBuild, buildID, filter)
	it := client.ListBuilds(ctx, &cloudbuildpb.ListBuildsRequest{
		Parent: "projects/" + projectID + "/locations/" + location,
		Filter: filter,
	})
	for {
		inProgressBuild, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
				return cli.Exit("error: filter is probably invalid: "+filter, 1)
			}
			return err
		}
		if _, err := client.CancelBuild(ctx, &cloudbuildpb.CancelBuildRequest{
			Name: "projects/" + projectID + "/locations/" + location + "/builds/" + inProgressBuild.GetId(),
		}); err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
				// this build is no longer in progress
				continue
			}
			return err
		}
		fmt.Fprintf(os.Stderr, "cancel: %s\n", inProgressBuild.GetLogUrl())
	}
	fmt.Fprintln(os.Stderr, "done")
	return nil
}
