####################################################################
#
# Go Forth Test Script
#
####################################################################

"Hello from Go Forth!" .green
"====================" .green 
cr

############################################################

"Hill'o Stars" .green
1 5 .. {hill} each

############################################################

"Reversing (version 1) ..." .green
[1 2 [3 3 3] 4 5 6 [7 7] 8 9 10] rreverse . cr

"Iterative Reversing ..." .purple
{1 30 .. nrev} duration . cr

############################################################
#
# List reversal using 'linrec'
#
DEFINE linrev == {len 2 < } {} {uncons} {swap append} linrec;

"Reversing (version 4 with 'linrec') ..." .blue
{1 30 .. linrev} duration .yellow cr

############################################################
"Flattening a list ..." .green
[1 2 [3 3 3] 4 5 6 [7 7] 8 9 10] list:flatten . cr

############################################################
"Fibonacci computation..." .green
{
    "Fib(25)=" 25 fib + .yellow
}
duration
cr

"Fibonacci computation using 'binrec' ..." .green
{
    "Fib(25)=" 25 bfib + .yellow
}
duration
cr

"Iterative Fibonacci computation ..." .green
{
    "Fib(25)=" 25 ifib + .yellow
}
duration
cr

"Point-free Iterative Fibonacci computation ..." .green
{
    "Fib(25)=" 25 pfib + .yellow
}
duration
cr

############################################################
#
# Filter out odd numbers
#
"Filtering with 'filter'..." .green
1 50 .. {2 % 0 ==} filter .yellow
cr

"Quick Sorting..." .green
{
     50 list:random {100 %} map qsort .yellow
}
duration
cr

"Quick Sorting using 'binrec' ..." .green
{
     50 list:random {100 %} map bqsort .yellow
}
duration
cr

############################################################
#
# Filtering a list using primrec
#
"Filtering using primrec ..." .blue
1 20 .. {[]} {first dup 2 % 0 == {append} {pop} ifte} primrec .
cr

"Factorials using primrec ..." .blue
1 10 .. {{1} {*} primrec} map .

############################################################
#
# Filtering lines in a file to get all the 'DEFINE's.
#
"Getting defined functions in the prelude..." .blue
"prelude.gf" file:read                  # read the file text
    r/\n/ str:split                     # split it into an array of lines
        r/^DEFINE/ str:match            # only keep lines starting with /^DEFINE/
            { r/ +/ str:split 1 @}    # Print out each result
            map
                .
cr

############################################################

"Executing fizbuz..." .blue
1 20 .. {fizbuz} map .
cr

############################################################
#
# Functions for error message testing
#
DEFINE fabble ==
    "Fabbling!" .yellow
    cstk
    boom
;

"Defining fibble!" .cyan
DEFINE fibble ==
    "Fibbling!" .green
    cstk
    fabble
;

############################################################
"Print a binary tree" .blue
[1
    [2 nil
        [6 nil
            [7 nil
                [9 nil nil]]]]
    [3 nil 
        [4
            [5 nil nil]
            nil]]]
                ptree


############################################################
#
# List of factorials of numbers from 1 10
#
"Compute factorials with 'linrec':" .blue
1 10 .. {{2 <} {pop 1} {dup 1 -} {*} linrec} map .blue
cr

############################################################
#
# List of nominal functions in the go source file
#
"List of nominal functions defined in goforth.go..." .blue
"goforth.go" file:readlines
    r/^ *func / str:match
        r/^func (\([^\)]*\))? *([^ ]+)\(.*$/ "$2" str:replace
            {.yellow} each
cr

"All Done - Bye!" .green

