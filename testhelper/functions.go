package testhelper

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/adam8157/gp-common-go-libs/dbconn"
	"github.com/adam8157/gp-common-go-libs/gplog"
	"github.com/adam8157/gp-common-go-libs/operating"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

/*
 * Functions for setting up the test environment and mocking out variables
 */

func SetupTestLogger() (*gbytes.Buffer, *gbytes.Buffer, *gbytes.Buffer) {
	testStdout := gbytes.NewBuffer()
	testStderr := gbytes.NewBuffer()
	testLogfile := gbytes.NewBuffer()
	testLogger := gplog.NewLogger(testStdout, testStderr, testLogfile, "gbytes.Buffer", gplog.LOGINFO, "testProgram")
	gplog.SetLogger(testLogger)
	return testStdout, testStderr, testLogfile
}

func SetupTestEnvironment() (*dbconn.DBConn, sqlmock.Sqlmock, *gbytes.Buffer, *gbytes.Buffer, *gbytes.Buffer) {
	testStdout, testStderr, testLogfile := SetupTestLogger()
	connection, mock := CreateAndConnectMockDB(1)
	operating.System = operating.InitializeSystemFunctions()
	return connection, mock, testStdout, testStderr, testLogfile
}

func CreateMockDB() (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	mockdb := sqlx.NewDb(db, "sqlmock")
	if err != nil {
		Fail("Could not create mock database connection")
	}
	return mockdb, mock
}

func SetDBVersion(connection *dbconn.DBConn, versionStr string) {
	connection.Version = dbconn.GPDBVersion{VersionString: versionStr, SemVer: semver.MustParse(versionStr)}
}

func CreateAndConnectMockDB(numConns int) (*dbconn.DBConn, sqlmock.Sqlmock) {
	mockdb, mock := CreateMockDB()
	driver := TestDriver{DB: mockdb, DBName: "testdb", User: "testrole"}
	connection := dbconn.NewDBConnFromEnvironment("testdb")
	connection.Driver = driver
	connection.MustConnect(numConns)
	SetDBVersion(connection, "5.1.0")
	return connection, mock
}

func ExpectRegexp(buffer *gbytes.Buffer, testStr string) {
	Expect(buffer).Should(gbytes.Say(regexp.QuoteMeta(testStr)))
}

func NotExpectRegexp(buffer *gbytes.Buffer, testStr string) {
	Expect(buffer).ShouldNot(gbytes.Say(regexp.QuoteMeta(testStr)))
}

func ShouldPanicWithMessage(message string) {
	if r := recover(); r != nil {
		errorMessage := strings.TrimSpace(fmt.Sprintf("%v", r))
		if !strings.Contains(errorMessage, message) {
			Fail(fmt.Sprintf("Expected panic message '%s', got '%s'", message, errorMessage))
		}
	} else {
		Fail("Function did not panic as expected")
	}
}

func AssertQueryRuns(connection *dbconn.DBConn, query string) {
	_, err := connection.Exec(query)
	Expect(err).To(BeNil(), "%s", query)
}

/*
 * This function call should be followed by a call to InitializeSystemFunctions
 * in a defer statement or AfterEach block.
 */
func MockFileContents(contents string) {
	r, w, _ := os.Pipe()
	operating.System.OpenFileRead = func(name string, flag int, perm os.FileMode) (operating.ReadCloserAt, error) { return r, nil }
	_, _ = w.Write([]byte(contents))
	_ = w.Close()
}
