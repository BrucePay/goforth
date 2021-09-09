####################################################################
#
# Go Forth Prelude
#
####################################################################


DEFINE ansi:black  == "\e[30m";
DEFINE ansi:red    == "\e[31m";
DEFINE ansi:green  == "\e[32m";
DEFINE ansi:yellow == "\e[33m";
DEFINE ansi:blue   == "\e[34m";
DEFINE ansi:purple == "\e[35m";
DEFINE ansi:cyan   == "\e[36m";
DEFINE ansi:white  == "\e[37m";

############################################################
#
# Screen control functions
#
DEFINE cls            == "\e[H\e[0J" print;
DEFINE console:home   == "\e[H" print;
DEFINE console:red    == ansi:red print;
DEFINE console:green  == ansi:green print;
DEFINE console:yellow == ansi:yellow print;
DEFINE console:blue   == ansi:blue print;
DEFINE console:purple == ansi:purple print;
DEFINE console:cyan   == ansi:cyan print;
DEFINE console:white  == ansi:white print;
DEFINE console:black  == ansi:black print;
DEFINE console:reset  == "\e[0m" print;


############################################################
#
# Function to print a carriage return
#
DEFINE cr == "" .;

############################################################
#
# Function to print a line of stars
#
DEFINE stars == "*" swap * .yellow;

############################################################
#
# Function to print a "hill" of stars
#
DEFINE hill ==
    dup { $_ stars } repeat
    dup { dup $_ - stars } repeat
    pop
;

############################################################
#
# Define a function to compute the execution duration of a lambda
#
DEFINE duration : start_time ==
    datetime -> start_time
    &
    "Duration " start_time since + .cyan
;

############################################################
#
# Define the zip function (uses variables).
# array1 array2 -> resultArray
#
DEFINE zip l1 l2 prog : r x y result ==
    [] -> result                    # initialize the result vector
    {l1 empty? l2 empty? or not?} # make sure neither array is empty
    {
        l1 uncons -> l1 -> x        # get the head of each collection
        l2 uncons -> l2 -> y
        x y prog &  -> r            # apply the specified program
        result r + -> result        # add the result to the result list
    }
    while
    result                          # and return the result
    nil -> result
;

############################################################
#
# DEFINE a function to recursively reverse a list
#
DEFINE rreverse ==
    uncons dup empty?
    {pop [] swap +}
    {rreverse swap +}
    ifte
;

############################################################
#
# Using uncons to simplify reverse
#
DEFINE rreverse2 ==
    uncons
    dup empty?
    {pop [] swap +}
    {rreverse2 swap +}
    ifte
;

############################################################
#
# Iterative list reversal
#
DEFINE nrev ==
    [] swap
    {dup}
    {uncons {swap cons} dip}
    while
    pop
;

############################################################
#
# List reversal using 'linrec'
#
DEFINE reverse ==
    {small}         # if the arg list is small
    {}              # just return it
    {uncons}        # otherwise uncons and recurse
    {swap append}   # join everything back up again
    linrec
;

############################################################
#
# Iterative list reversal
#
DEFINE ireverse lst : x x1 y y1 ==
    # initialize the start and end indexes
    0 -> x
    lst len pred -> y
    # while x is < y loop...
    {x y <}
    {
        lst x @             # get the start index value
        lst y @             # get the end index value
        lst swap x swap !   # store the values reversed
        lst swap y swap !
        x succ -> x         # increment the indexes
        y pred -> y
    }
    while
    lst                 # return the updated list
;

############################################################
#
# Function to cause a runtime error
#
DEFINE boom ==
    # ijijiljljlkjlkjlk
    # barf # calling a function that does not exist
    # iuhjujljljlkjlkj
    1 0 / # divide by zero
    # all done
;

############################################################
#
# Define a recursive point-free Fibonacci function
#
DEFINE fib n ==
    n 2 <                    # if less than 2
    {1}                      # return 1
    {n pred fib n 2 - fib +}  # otherwise compute fib(n-1) + fib(n-2)
    ifte
;

############################################################
#
# Define a Fibonacci function using 'binrec'
#
DEFINE bfib ==
    {2 <}                         # if less than 2
    {pop 1}                       # pop the arg and return 1
    {pred dup pred}                 # otherwise recurse with n-1 and n-2
    {+}                           # sum the results
    binrec
;

############################################################
#
# Define an iterative Fibonacci function (uses variables)
#
DEFINE ifib num : current next ==
    0 -> current
    1 -> next
    {num}                          # while the number is not zero
    {
        next next current +        # compute the new current and next values
            -> next -> current    
        num pred -> num
    }
    while
    next                           # return the final result
;

############################################################
#
# Define a point-free iterative Fibonacci function
#
DEFINE pfib ==
    [0 1]                               # initialize the sequence [curr next]
    {{pred} dip over}                    # check remaining iteration
    {dup 1 @ swap first over + append}  # update the sequence [next curr+next]
    while
    swap pop last                       # get rid of the count and return
;

############################################################
#
# Define a factorial function using linrec
#
DEFINE fact ==
    {small} {pop 1}     # if the arg is < 2 return 1
    {dup pred} {*}      # multiply the argument and it's predecessor
    linrec
;

############################################################
#
# Recursive quicksort implementation
#
DEFINE qsort ==
    dup
    {
        uncons {over >} list:split  # split the list into pivot, smaller and larger
        dup {qsort} if              # if not empty, sort the smaller list
        swap rol append             # append the pivot to the smaller list
        swap
        dup {qsort} if              # if not empty, sort the larger list
        swap append                 # append the larger list to the smaller
    }
    if
;

############################################################
#
# Quicksort implementation using 'binrec'
#
DEFINE bqsort ==
    {small}                         # if the list is small
    {}                              # just return it
    {uncons {over <} list:split}    # otherwise split into pivot, smaller and larger and recurse
    {swapd cons append}             # recombine the elements in order
    binrec
;

############################################################
#
# Define the zero? predicate - returns true if the argument is zero
#
DEFINE zero? == 0 ==;

############################################################
#
# Fizbuz implementation using the 'case' function
#
DEFINE fizbuz ==
    [
        {15 % zero?} "fizbuz"
        {3  % zero?} "fiz"
        {5  % zero?} "buz"
        {pop true}   {}
    ]
    case
;

############################################################
#
# Print a binary tree where nodes are lists of three items e.g. [val left right]
#
DEFINE ptree tree ==
    tree notempty?          # if the tree isn't empty
    {
       tree 0 @ .yellow     # print the value (1st element)
       tree 1 @ ptree       # recurse and print the left tree
       tree 2 @ ptree       # recurse print the right tree
    }
    if
;

############################################################
#
# Return the minimum of 2 values
#
DEFINE min ==
    # find the smallest of two values
    dup2 > {swap pop} {pop} ifte
;

############################################################
#
# Return the max of 2 values
#
DEFINE max ==
    # find the largest of two values
    dup2 < {swap pop} {pop} ifte
;

############################################################
#
# Return the minimum value in a list
#
DEFINE list:min == {min} reduce;

############################################################
#
# Return the maximum value in a list
#
DEFINE list:max == {max} reduce;

############################################################
#
# Return the sum of all of the values in a list
#
DEFINE list:sum == {+} reduce;

############################################################
#
# Return the product of all of the values in a list
#
DEFINE list:prod == {*} reduce;

############################################################
#
# Return the average of all of the values in a list
#
DEFINE list:average == {list:sum} {len} cleave /;

############################################################
#
# Flatten a list using linrec
#
DEFINE list:flatten ==
    {empty?}                        # if the list is empty, return []
    {}
    {uncons swap dup list? {list:flatten} if swap} # otherwise flatten the list one at a time recursively
    {append}
    linrec
;

############################################################
#
# Return all of the primes up to a specific number
#
DEFINE primes : r v ==
    [] -> r                     # initialize the result collection variable
    3 swap 2 ...                # generate 3, 5, 7, ... n
    {dup notempty?}             # while the list is not empty
    {
        uncons swap dup r swap + -> r -> v {v % true?} filter
    }
    while
    pop
    r
;

############################################################
#
# Pad a string out to a certain length on the left side
#
DEFINE str:padleft ==
   swap dup len rol rol swap - " " swap * swap +
;

############################################################
#
# Pad a string out to a certain length on the right side
#
DEFINE str:padright ==
    swap dup len rol rol swap - " " swap * +
;

############################################################
#
# Print out all of the currently defined words formatted in groups of 6 padded to 20 spaces
#
DEFINE help ==
    console:yellow
    ops keys                            # get the names of all of the defined operations
        sort                            # sort the names
            6 list:split                # split into groups of 6
            {
                {20 str:padright} map   # pad each field in the group
                str:join .              # join the group into a string and print it
            }
            each
    console:reset
;

############################################################
#
# Print out detailed help for all of the currently defined words
#
DEFINE help:detailed : helpComments helpMap key ==
    "" -> helpComments
    [] dict! -> helpMap
    "goforth.go" file:readlines
        {
            [
                r/ *\/\/C/ {
                    4 skip helpComments swap + "\n" + -> helpComments
                }
                r/ops.".*= / {
                    r/^.*ops."([^"]+)".*$/ "$1" str:replace
                        r/^[ \t\r\n]+/ "" str:replace -> key
                    helpMap key helpComments !
                    "" -> helpComments
                }
            ] case
         }
         each
    "==================== Help ==============================" .green
    helpMap keys sort {dup .green helpMap swap @ .yellow} each
;

############################################################
#
# Compute the greatest common divisor using Euclid's method.
#
DEFINE gcd ==
    dup 0 ==
    {pop}
    {
        dup     # x y y 
        rol     # y x y
        %       # y x%y
        gcd
    }
    ifte
;

############################################################
#
# Compute the GCD using Euclid's method and 'linrec'
#
DEFINE gcd2 ==
    {0 ==}          # if the second arg is zero
    {pop}           # pop it and return the first arg
    {dup rol %}     # compute the next pair of values
    {}              # end block is a noop in this case
    linrec
;


