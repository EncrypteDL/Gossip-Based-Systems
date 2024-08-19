package registry

import "os"

const statusFile = "./experiment_status.txt"

type StatusLogger struct{
	file *os.File
}

func NewStatutsLogger() *StatusLogger{
	file, err := os.OpenFile(statusFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil{
		panic(err)
	}

	statusLogger := &StatusLogger{
		file: file,
	}

	return statusLogger
}

func (s *StatusLogger) LogStarted() {
	s.file.WriteString("started\n")
	s.file.Sync()
}

func (s *StatusLogger) LogFailed() {
	s.file.WriteString("failed\n")
	s.file.Sync()
}

func (s *StatusLogger) LogFinished() {
	s.file.WriteString("completed\n")
	s.file.Sync()
}