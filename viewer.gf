#########################################################
#
# Display the argument file on the console.
#
# Takes one argument: the name of the file to view.
#
#########################################################

file:readlines -> lines   # read the lines into a variable

0    -> cl                # set up a line counter
true -> loop              # set up the command loop flag
40   -> linesPerPage      # set up the number of lines per page

#########################################################
#
# Define the help text
#
"
-----------------------------------------------------------------------------
Available Commands:

?       - display help
q       - quit the viewer
u       - move up one page
d       - move down half a page
k       - move up one line
j       - move down one line
/<pat>  - search forward for a line matching the <pat> regular expression
<num>   - go to line <num>
-<num>  - go backwards <num> lines
+<num>  - go forwards <num> lines
<cr>    - move down a page

Press enter to return to the viewer.

"
-> helpText

#########################################################

{cl lines len < loop and}      # while we haven't shown all the lines
{

    # clear the screen then render the page
    cls
    0 linesPerPage .. {
        pop
        # print line number in green
        console:green
        "%-5s | " [cl int! string!] format print
        console:reset
        lines cl @ .  # print the line in the default color
        cl 1 + -> cl
    }
    each

    # prompt for and process the user commands
    "      -----------------------------------------------------------------------------" .green
    console:cyan
    "Press enter to continue, u<enter> to move up, q<enter> to quit.\n> " print
    console:reset
    getline
    str:trim
    [
        # search forwards for the specified text
        r/^\/./   {
                        1 skip regex! -> pat
                        "Searching for " pat + .yellow
                        cl -> fl
                        true -> search
                        {fl lines len < search and}
                        {
                            lines fl @ pat str:match
                            {
                                fl -> cl
                                false -> search
                            }
                            {
                                fl 1 + -> fl
                            }
                            ifte
                        }
                        while
                    }

        # Goto the specified line
        r/^[0-9]+/   {
                        -> lineToGoTo
                        lineToGoTo
                        {
                            lineToGoTo int!
                            [
                                {0 <} {pop 0}
                                {lines len >=} {pop lines len linesPerPage -}
                                {pop true} {}
                            ]
                            case
                            int!
                            -> cl
                        }
                        if
                    }

        # Goto the - relative line
        r/^-[0-9]+/   {
                        r/-([0-9]+)/ "$1" str:replace int! -> lineToGoTo
                        lineToGoTo
                        {
                            cl linesPerPage - lineToGoTo - pred -> cl
                            cl 0 < { 0 -> cl } if
                        }
                        if
                    }

        # Goto the + relative line
        r/^\+[0-9]+/   {
                        r/\+([0-9]+)/ "$1" str:replace int! -> lineToGoTo
                        lineToGoTo
                        {
                            cl linesPerPage - lineToGoTo + -> cl
                            cl 0 < { 0 -> cl } if
                        }
                        if
                    }

        # Move up 1 line
        r/^[Kk]/     {
                        pop
                        cl linesPerPage - 2 - -> cl
                        cl 0 < {0 -> cl} if
                    }

        # Move down 1 line
        r/^[jJ]/      {
                        pop
                        cl linesPerPage - -> cl
                    }

        # Quit the viewer
        r/^[qQ]/     {
                        pop
                        false -> loop
                    }

        # move up one page
        r/^[uU]/     {
                        pop
                        cl linesPerPage 2 * - -> cl
                        cl 0 < {0 -> cl} if
                    }

        # move down half a page
        r/^[dD]/     {
                        pop
                        cl linesPerPage 2 / - -> cl
                    }

        r/\?/       {
                        pop
                        helpText .yellow
                        getline
                    }


        # move to the next page
        r/^ */      {pop}
    ]
    case
}
while
nil -> lines


