package remote_shell

import "io"

type Command struct {
	Input  io.WriteCloser
	Output io.ReadCloser
	cmd Cmd
}



