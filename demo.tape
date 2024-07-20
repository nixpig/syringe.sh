Output demo.gif

Require syringe

Set FontSize 28
Set FontFamily "Operator Mono Lig"
Set Theme "Catppuccin Mocha"
Set Width 1200
Set Height 600

# register
Hide
Type "# register a user"
Enter
Show
Sleep 500ms
Type "syringe user register"
Sleep 250ms
Enter
Sleep 2s

# add project
Hide
Type "clear"
Enter
Type "# add a project"
Enter
Show
Sleep 500ms
Type "syringe project add my_cool_project"
Sleep 250ms
Enter
Sleep 2s

# add environment
Hide
Type "clear"
Enter
Type "# add an environment"
Enter
Show
Sleep 500ms
Type "syringe environment add -p my_cool_project local"
Sleep 250ms
Enter
Sleep 2s

# set secret
Hide
Type "clear"
Enter
Type "# set a secret"
Enter
Show
Sleep 500ms
Type "syringe secret set -p my_cool_project -e local SUPER_SECRET shhh!"
Sleep 250ms
Enter
Sleep 2s

# get secret
Hide
Type "clear"
Enter
Type "# get a secret"
Enter
Show
Sleep 500ms
Type "syringe secret get -p my_cool_project -e local SUPER_SECRET"
Sleep 250ms
Enter
Sleep 3.5s

# set another secret
Hide
Type "clear"
Enter
Type "# set another secret..."
Enter
Show
Sleep 500ms
Type 'syringe secret set -p my_cool_project -e local MORE "another secret"'
Sleep 250ms
Enter
Sleep 250ms

# set yet another secret
Hide
Enter
Type "# ...and another"
Enter
Show
Sleep 250ms
Type 'syringe secret set -p my_cool_project -e local MOAARRRR "here we go!"'
Sleep 250ms
Enter
Sleep 2s

# list secrets
Hide
Type "clear"
Enter
Type "# list secrets"
Enter
Show
Sleep 250ms
Type "syringe secret list -p my_cool_project -e local"
Sleep 250ms
Enter
Sleep 2s

# inject secrets
Hide
Type "clear"
Enter
Type "# inject secrets into a command/script"
Enter
Show
Sleep 250ms
Type "syringe secret inject -p my_cool_project -e local -- scripts/test.sh"
Sleep 250ms
Enter
Sleep 3s

# clear down
Hide
Type "clear"
Enter
Show
Sleep 2s
