package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CommandEngine - Represent struct of command engine
type CommandEngine struct {
	rootCmd *cobra.Command
}

func NewCommand(name string) *CommandEngine {
	c := &cobra.Command{
		Use:   name,
		Short: "Mocking Server for simulate biller request and response",
		Long: `Mocking Biller is a service to simulate Request and Response to and from biller, 
			   simulate Callback and give response`,
	}

	defer func() {
		r := recover()
		if r != nil {
			logrus.Error(r)
		}
	}()

	c.PersistentFlags().StringP("config", "c", "configurations", "the config path location")

	return &CommandEngine{
		rootCmd: c,
	}
}

// GetRoot - Get Application Root
func (c *CommandEngine) GetRoot() *cobra.Command {
	return c.rootCmd
}

// Run - Run Command
func (c *CommandEngine) Run() {
	tcpiso := &cobra.Command{
		Use:   "iso",
		Short: "ISO Server Mocking",
		Long:  "ISP Server for simulate third party TCP with ISO",
		Run: func(*cobra.Command, []string) {
			fmt.Println("Run iso server in port:", "8091")
			StartServerMode()
		},
	}

	var commands = []*cobra.Command{tcpiso}

	for _, command := range commands {
		c.rootCmd.AddCommand(command)
	}

	err := c.rootCmd.Execute()
	if err != nil {
		fmt.Printf("error while executing root command, got: %v", err)
	}
}
