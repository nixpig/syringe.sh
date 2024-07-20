Output demo.gif

Require syringe

Set FontSize 28
Set Width 1200
Set Height 600

Type "# register a user"
Enter
Type "syringe user register"
Sleep 250ms
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "# add a project"
Enter
Type "syringe project add my_cool_project"
Sleep 250ms
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "# add an environment"
Enter
Type "syringe environment add -p my_cool_project local"
Sleep 250ms
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "# set a secret"
Enter
Type "syringe secret set -p my_cool_project -e local SUPER_SECRET shhh!"
Sleep 250ms
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "# get a secret"
Enter
Type "syringe secret get -p my_cool_project -e local SUPER_SECRET"
Sleep 250ms
Enter
Sleep 3.5s

Hide
Type "clear"
Enter
Show

Type "# set another secret..."
Enter
Type 'syringe secret set -p my_cool_project -e local MORE "another secret"'
Sleep 250ms
Enter
Sleep 250ms

Type "# ...and another"
Enter
Type 'syringe secret set -p my_cool_project -e local MOAARRRR "here we go!"'
Sleep 250ms
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "# list secrets"
Enter
Type "syringe secret list -p my_cool_project -e local"
Sleep 250ms
Enter
Sleep 2s

Hide
Type "clear"
Enter
Show

Type "# inject secrets into a command/script"
Enter
Type "syringe secret inject -p my_cool_project -e local -- scripts/test.sh"
Sleep 250ms
Enter
Sleep 2s

Sleep 3s
