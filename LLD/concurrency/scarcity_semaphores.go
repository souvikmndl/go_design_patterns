package concurrency

import (
	"context"

	"golang.org/x/sync/semaphore"
)

/*
A semaphore is like a bouncer at a club whose sole job
is to make sure that the number of people inside dont exceed
capacity.
In our example, we can make only 5 APIClient requests at a time
That is why we initialise the semaphore with capacity 5
like: semaphore.NewWeighted(5)
Everytime MakeRequest() is called, thread will try to get a semaphore
first by calling Acquire(). If it fails, that means capacity is reached
for now and error is returned. Else, it gets a lock and count internally
reduces by 1. Upon completion, we release that semaphore and count
increases by 1 again.
It is basically like a rate limiter

However, keep in mind that Semaphores are just there to limit the
NUMBER of certain actions being performed concurrently.
If you need to habnd out  resources, then pooling is the solution
*/

type APIClient struct {
	sem        *semaphore.Weighted
	httpClient httpClient
}

type Response struct{}
type httpClient struct{}

func (c *httpClient) Get(endpoint string) (Response, error) {
	return Response{}, nil
}

func NewAPIClient() *APIClient {
	return &APIClient{sem: semaphore.NewWeighted(5)}
}

func (c *APIClient) MakeRequest(ctx context.Context, endpoint string) (Response, error) {
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return Response{}, err
	}
	defer c.sem.Release(1)
	return c.httpClient.Get(endpoint)
}
