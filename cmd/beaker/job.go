package main

import (
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
)

// listJobs follows all cursors to get a complete list of jobs.
func listJobs(opts client.ListJobOpts) ([]api.Job, error) {
	var jobs []api.Job
	for {
		var page api.Jobs
		page, err := beaker.ListJobs(ctx, &opts)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, page.Data...)
		if page.Next == "" {
			break
		}
		opts.Cursor = page.Next
	}
	return jobs, nil
}
