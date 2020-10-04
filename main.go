// Copyright 2020 Marius van der Wijden
// This file is part of the fuzzy-vm library.
//
// The fuzzy-vm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The fuzzy-vm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the fuzzy-vm library. If not, see <http://www.gnu.org/licenses/>.

// Package main creates a fuzzer for Ethereum Virtual Machine (evm) implementations.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/MariusVanDerWijden/FuzzyVM/executor"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = "FuzzyVM"
	app.Author = "Marius van der Wijden"
	app.Usage = "Generator for Ethereum Virtual Machine tests"
	app.Action = mainLoop
	app.Flags = []cli.Flag{
		genProcFlag,
		maxTestsFlag,
		minTestsFlag,
		buildFlag,
	}
	return app
}

var app = initApp()

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainLoop(c *cli.Context) {
	if c.GlobalBool(buildFlag.Name) {
		if err := startBuilder(); err != nil {
			panic(err)
		}
	} else {
		generatorLoop(c)
	}
}

func generatorLoop(c *cli.Context) {
	var (
		genProc  = c.GlobalString(genProcFlag.Name)
		minTests = c.GlobalInt(minTestsFlag.Name)
		maxTests = c.GlobalInt(maxTestsFlag.Name)
	)
	for {
		fmt.Println("Starting generator")
		errChan := make(chan error)
		cmd := startGenerator(genProc)
		go startExecutor(errChan)
		go watcher(cmd, errChan, maxTests)

		err := <-errChan
		if err != nil {
			panic(err)
		}
		infos, err := ioutil.ReadDir("out")
		if err != nil {
			panic(err)
		}
		if len(infos) > minTests {
			fmt.Println("Tests exceed minTests after execution")
			return
		}
	}
}

func startBuilder() error {
	cmdName := "go-fuzz-build"
	cmd := exec.Command(cmdName)
	cmd.Dir = "fuzzer"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// We have to disable CGO
	cgo := "CGO_ENABLED=0"
	env := append(os.Environ(), cgo)
	cmd.Env = env
	return cmd.Run()
}

func startGenerator(genThreads string) *exec.Cmd {
	cmdName := "go-fuzz"
	dir := "./fuzzer/fuzzer-fuzz.zip"
	cmd := exec.Command(cmdName, "--bin", dir, "--procs", genThreads)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	return cmd
}

func startExecutor(errChan chan error) {
	errChan <- executor.Execute("out", "crashes")
}

func watcher(cmd *exec.Cmd, errChan chan error, maxTests int) {
	for {
		time.Sleep(time.Second * 5)
		infos, err := ioutil.ReadDir("out")
		if err != nil {
			fmt.Printf("Error killing process: %v\n", err)
			cmd.Process.Kill()
			errChan <- err
		}
		if len(infos) > maxTests {
			fmt.Printf("Max tests exceeded, pausing")
			cmd.Process.Signal(os.Interrupt)
			return
		}
	}
}
