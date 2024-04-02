package main

import (
	"errors"
	"github.com/Neutaro/Neutaro-Chain/app/params"
	"github.com/Neutaro/Neutaro-Chain/cmd/Neutaro/cmd"
	"github.com/cosmos/cosmos-sdk/server"
	"os"

	"github.com/Neutaro/Neutaro-Chain/app"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	params.SetAddressPrefixes()
	params.RegisterDenoms()

	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		var e server.ErrorCode
		switch {
		case errors.As(err, &e):
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}
