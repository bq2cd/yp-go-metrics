// Binary keygen generates X25519 key pair encoded in ASN.1 DER form.
// The generated keys are printed to standard output or written into files
// (depending on CLI flags).
package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"os"

	"github.com/bq2cd/yp-go-metrics/internal/app/server/servertest"
)

var (
	flagOutPath = flag.String("o", "", "Output path for private key (public key will get .pub extension)")
)

type writerPair struct {
	public  io.Writer
	private io.Writer
}

func main() {
	flag.Parse()

	keys, err := servertest.NewX25519KeyPair()
	if err != nil {
		log.Fatalln(err)
	}

	err = printOrWriteToFile(keys, *flagOutPath)
	if err != nil {
		log.Fatalln(err)
	}
}

func printOrWriteToFile(keys *servertest.X25519KeyPair, outPath string) (err error) {
	var (
		writers *writerPair
		closeFn func() error
	)

	writers, closeFn, err = getWriters(outPath)
	if err != nil {
		return
	}

	defer func() {
		err = errors.Join(err, closeFn())
	}()

	err = errors.Join(
		func() error {
			_, err = keys.Private.WriteTo(writers.private)

			return err
		}(),
		func() error {
			_, err = keys.Public.WriteTo(writers.public)

			return err
		}(),
	)

	return
}

func getWriters(outPath string) (*writerPair, func() error, error) {
	wpair := &writerPair{
		private: os.Stdout,
		public:  os.Stdout,
	}

	if outPath == "" {
		return wpair, func() error { return nil }, nil
	}

	var (
		privateFile *os.File
		publicFile  *os.File
		err         error
	)

	privateFile, err = os.Create(outPath)
	if err != nil {
		return nil, nil, err
	}

	publicFile, err = os.Create(outPath + ".pub")
	if err != nil {
		return nil, nil, err
	}

	wpair.private = privateFile
	wpair.public = publicFile

	closeFn := func() error {
		return errors.Join(
			privateFile.Close(),
			publicFile.Close(),
		)
	}

	return wpair, closeFn, nil
}
