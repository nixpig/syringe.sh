Output demo.gif

Require syringe
Require starship

Set FontSize 28
Set FontFamily "Operator Mono Lig"
Set Theme "Catppuccin Mocha"
Set Width 1200
Set Height 600
Set WindowBar Rings
Set BorderRadius 10

# register
Show
Sleep 500ms
Type "syringe register"
Sleep 150ms
Enter
Sleep 2s

Type "syringe set foo bar"
Sleep 150ms
Enter
Sleep 1s

Type "syringe set baz qux"
Sleep 150ms
Enter
Sleep 1s

Type "syringe set hur dur"
Sleep 150ms
Enter
Sleep 2s

Hide
Type "PROMPT_COMMAND=echo"
Enter
Type "clear"
Enter
Show
Sleep 1s

Type "syringe list"
Sleep 150ms
Enter
Sleep 2s

Type "syringe get baz"
Sleep 150ms
Enter
Sleep 2s

Type "syringe get hur"
Sleep 150ms
Enter
Sleep 2s

Hide
Type "PROMPT_COMMAND="
Enter
Type "clear"
Enter
Show

Type "syringe remove baz"
Sleep 150ms
Enter
Sleep 1s

Type "syringe remove foo"
Sleep 150ms
Enter
Sleep 1s

Hide
Type "PROMPT_COMMAND=echo"
Enter
Type "clear"
Enter
Show

Type "syringe list"
Sleep 150ms
Enter
Sleep 2s


# clear down
Hide
Type "clear"
Enter
Show
Sleep 2s
