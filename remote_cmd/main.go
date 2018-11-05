package main

import (
	"os/exec"
	"os"
	"fmt"
		"bufio"
	"flag"
)

func main() {

	//ini_path := flag.String("c", defaultConfigFilePath, "指定配置文件路径")
	old_config := flag.String("oc" , "" , "指定老的配置文件" )

	flag.Parse()
	version := "0.0.3"
	fmt.Printf("version : %s\n", version)
	fmt.Printf("config path : %s\n", *old_config)


	cmd := exec.Command("bash")
	output, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error ", err)
		return
	}
	defer output.Close()

	input, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error ", err)
		return
	}
	defer input.Close()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for ; ; {
			result, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("read error:", err)
				return
			}
			fmt.Print("input:", result)

			_, err = input.Write([]byte(result))
			if err != nil {
				fmt.Print("input.Write:", err)
			}
		}
	}()

	go func() {
		reader := bufio.NewReader(output)
		for ; ; {
			result, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("read error:", err)
				return
			}
			fmt.Print("output:", result)
		}
	}()

	cmd.Start()

}
