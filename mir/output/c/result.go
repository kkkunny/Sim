package c

import (
	"os"
	"os/exec"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

type Result struct {
	code strings.Builder
}

func (self Result) String()string{
	return self.code.String()
}

func (self Result) Output(out stlos.FilePath)stlerror.Error{
	file, err := stlerror.ErrorWith(os.OpenFile(out.String(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666))
	if err != nil{
		return err
	}
	defer file.Close()

	_, err = stlerror.ErrorWith(file.WriteString(self.code.String()))
	return err
}

func (self Result) Build(out stlos.FilePath)stlerror.Error{
	path, err := stlerror.ErrorWith(stlos.RandomTempFilePath("sim"))
	if err != nil{
		return err
	}
	path = stlos.FilePath(strings.TrimSuffix(path.String(), path.Ext()) + ".c")

	err = self.Output(path)
	if err != nil{
		return err
	}
	defer os.Remove(path.String())

	if !out.IsAbs(){
		workPath, err := stlerror.ErrorWith(stlos.GetWorkDirectory())
		if err != nil{
			return err
		}
		out = workPath.Join(out.String())
	}
	cmd := exec.Command("gcc", "-std=c11", "-o", out.String(), path.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	originErr := cmd.Run()
	if originErr != nil && stlbasic.Is[*exec.ExitError](originErr){
		exitError := originErr.(*exec.ExitError)
		if exitError.ExitCode() != 0{
			return stlerror.Errorf(exitError.Error())
		}
	}else if originErr != nil{
		return stlerror.ErrorWrap(originErr)
	}
	return nil
}

func (self Result) Run()(uint8, stlerror.Error){
	path, err := stlerror.ErrorWith(stlos.RandomTempFilePath("sim"))
	if err != nil{
		return 0, err
	}
	path = stlos.FilePath(strings.TrimSuffix(path.String(), path.Ext()) + ".out")

	err = self.Build(path)
	if err != nil{
		return 0, err
	}
	defer os.Remove(path.String())

	cmd := exec.Command(path.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	originErr := cmd.Run()
	if originErr != nil && stlbasic.Is[*exec.ExitError](originErr){
		exitError := originErr.(*exec.ExitError)
		return uint8(exitError.ExitCode()), nil
	}else if originErr != nil{
		return 0, stlerror.ErrorWrap(originErr)
	}
	return 0, nil
}
