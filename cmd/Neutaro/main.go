package main

import (
	"errors"
	"os"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/Neutaro/Neutaro-Chain/app"
	"github.com/Neutaro/Neutaro-Chain/app/params"
	"github.com/Neutaro/Neutaro-Chain/cmd/Neutaro/cmd"
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
