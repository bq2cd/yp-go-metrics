// Package server provides main entry point and high level resources required to run the server application.
// These include launching HTTP server listener and a number of background goroutines responsible for
// dumping metrics to disk, writing metrics to database in batches, and others.
// The package also provides utility functions to apply database migration on server's startup.
package server
