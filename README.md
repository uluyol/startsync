# startsync

startsync waits for a provided number of clients to connect and
then broadcasts a message indicating that they should run.

startsync is useful for synchronizing the beginning of distributed
experiments. While startsync will not achieve exact
synchronization, in practice, it should be sufficient for tasks
like distributed load generation for benchmark purposes.

Docker image: https://hub.docker.com/r/uluyol/startsync/
