package queue

import (
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var testBase string

func TestMain(m *testing.M) {
	//setup()

	os.RemoveAll("../sandbox")
	os.Mkdir("../sandbox", 0755)
	var err error
	testBase, err = os.MkdirTemp("../sandbox", "testcases*")
	if err != nil {
		log.Fatalln(err)
	}
	err = cp.Copy("../testcases", testBase)
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()

	//shutdown()
	//os.RemoveAll(testBase)
	os.Exit(code)
}
