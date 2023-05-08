package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func CompleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "complete",
		Aliases:   []string{"c"},
		Usage:     "Generate shell completion scripts",
		ArgsUsage: "[SHELL]",
		Subcommands: []*cli.Command{
			{
				Name:        "bash",
				Action:      bashCompletion,
				Description: "generates a bash completion script",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "program",
						Usage: "Custom program name",
						Value: "harvester",
					},
				},
			},
			{
				Name:        "zsh",
				Action:      zshCompletion,
				Description: "generates a zsh completion script",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "program",
						Usage: "Custom program name",
						Value: "harvester",
					},
				},
			},
			{
				Name:        "powershell",
				Aliases:     []string{"ps"},
				Description: "generates a powershell completion script",
				Action:      powershellCompletion,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "program",
						Usage: "Custom program name",
						Value: "harvester",
					},
				},
			},
		},
	}
}

func bashCompletion(ctx *cli.Context) error {
	_, err := fmt.Printf(`#! /bin/bash

# This file is generated by 'harvester complete bash'.
# DO NOT EDIT THIS FILE

PROG=%s

: ${PROG:=$(basename ${BASH_SOURCE})}

# Macs have bash3 for which the bash-completion package doesn't include
# _init_completion. This is a minimal version of that function.
_cli_init_completion() {
  COMPREPLY=()
  _get_comp_words_by_ref "$@" cur prev words cword
}

_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base words
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if declare -F _init_completion >/dev/null 2>&1; then
      _init_completion -n "=:" || return
    else
      _cli_init_completion -n "=:" || return
    fi
    words=("${words[@]:0:$cword}")
    if [[ "$cur" == "-"* ]]; then
      requestComp="${words[*]} ${cur} --generate-bash-completion"
    else
      requestComp="${words[*]} --generate-bash-completion"
    fi
    opts=$(eval "${requestComp}" 2>/dev/null)
    COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
    return 0
  fi
}

complete -o bashdefault -o default -o nospace -F _cli_bash_autocomplete $PROG
`, ctx.String("program"))
	if err != nil {
		return err
	}
	return nil
}

func zshCompletion(ctx *cli.Context) error {
	_, err := fmt.Printf(`#compdef %s

_cli_zsh_autocomplete() {
local -a opts
local cur
cur=${words[-1]}
if [[ "$cur" == "-"* ]]; then
	opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
else
	opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
fi

if [[ "${opts[1]}" != "" ]]; then
	_describe 'values' opts
else
	_files
fi
}

compdef _cli_zsh_autocomplete %s
`, ctx.String("program"), ctx.String("program"))

	if err != nil {
		return err
	}

	return nil
}

func powershellCompletion(ctx *cli.Context) error {
	_, err := fmt.Printf(`$fn = $($MyInvocation.MyCommand.Name)
$name = $fn -replace "(.*)\.ps1$", '$1'
Register-ArgumentCompleter -Native -CommandName $name -ScriptBlock {
	param($commandName, $wordToComplete, $cursorPosition)
	$other = "$wordToComplete --generate-bash-completion"
		Invoke-Expression $other | ForEach-Object {
			[System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
`)

	if err != nil {
		return err
	}

	return nil
}
