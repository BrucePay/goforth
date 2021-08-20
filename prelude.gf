####################################################################
#
# Go Forth Prelude
#
####################################################################


############################################################
#
# Screen control functions
#
DEFINE cls            == "\e[H\e[0J" print;
DEFINE console:home   == "\e[H" print;
DEFINE console:red    == "\e[31m" print;
DEFINE console:green  == "\e[32m" print;
DEFINE console:yellow == "\e[33m" print;
DEFINE console:blue   == "\e[34m" print;
DEFINE console:purple == "\e[35m" print;
DEFINE console:cyan   == "\e[36m" print;
DEFINE console:white  == "\e[37m" print;
DEFINE console:reset  == "\e[0m" print;


############################################################
#
# Function to print a carriage return
#
DEFINE cr == "" .;

############################################################
#
# Return the predecessor of a value
#
DEFINE pred == 1 -;

############################################################
#
# Return the successor of a value
#
DEFINE succ == 1 +;

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
DEFINE duration ==
    datetime !_start_time
    &
    "Duration " $_start_time since + .cyan
;

DEFINE duration/2 ==
    datetime -> _start_time
    {dup -> iters} dip
    repeat

    "Duration " $_start_time since  + $iters / .cyan
;

############################################################
#
# Define the zip function (uses variables).
# array1 array2 -> resultArray
#
DEFINE zip ==
       -> _prog                         # save the zip program
    [] -> _result                       # initialize the result vector
    {dup2 {empty?} apply2 or not?}      # make sure neither array is empty
    {
        {uncons swap} apply2            # get the head of each collection
        rol rol swap                    # move the heads to the TOS
        $_prog &                        # apply the specified program
        $_result swap + -> _result      # add the result to the result list
    }
    while
    pop                                 # pop the empty arrays
    pop
    $_result                            # and return the result
    nil -> _result
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
    $dup
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
# Define a recursive Fibonacci function
#
DEFINE fib ==
    dup 2 <                         # if less than 2
    { pop 1 }                       # pop the arg and return 1
    {dup 1 - fib swap 2 - fib +}    # otherwise compute fib(n-1) + fib(n-2)
    ifte
;

############################################################
#
# Define a Fibonacci function using 'binrec'
#
DEFINE bfib ==
    {2 <}                         # if less than 2
    {pop 1}                       # pop the arg and return 1
    {1 - dup 1 -}                 # otherwise recurse with n-1 and n-2
    {+}                           # sum the results
    binrec
;

############################################################
#
# Define an iterative Fibonacci function (uses variables)
#
DEFINE ifib ==
    -> num                            # capture the number of iterations
    0 -> current                      # initialize current and next
    1 -> next
    {$num}                          # while the number is not zero
    {
        $next $next $current +      # compute the new current and next values
            -> next -> current    
        $num 1 - -> num
    }
    while
    $next                           # return the final result
;

############################################################
#
# Define a point-free iterative Fibonacci function
#
DEFINE pfib ==
    [0 1]                               # initialize the sequence [curr next]
    {{1 -} dip over}                    # check remaining iteration
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
        {}                          # just return it
    {uncons {over <} list:split}    # otherwise split into pivot, smaller and larger and recurse
        {swapd cons append}         # recombine the elements in order
    binrec
;

############################################################
#
# Fizbuz implementation using the 'case' operator
#
DEFINE fizbuz == [
        {15 % 0 ==} "fizbuz"
        {3 % 0 ==}  "fiz"
        {5 % 0 ==}  "buz"
        {pop true}  {}
    ]
    case
;

############################################################
#
# Print a binary tree where nodes are lists of three items e.g. [val left right]
#
DEFINE ptree == [
        nil {pop}
        {true!} {
            dup 0 @ .yellow
            dup 1 @ ptree
            2 @ ptree
        }
    ]
    case
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
    {uncons swap list:flatten swap} # otherwise flatten the list one at a time recursively
    {append}
    linrec
;

############################################################
#
# Return all of the primes up to a specific number
#
DEFINE primes ==
    [] -> r                     # initialize the result collection variable
    3 swap 2 ...                # generate 3, 5, 7, ... n
    {dup notempty?}           # while the list is not empty
    {
        uncons swap dup $r swap + -> r -> v {$v % true?} filter
    }
    while
    pop
    $r
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
                    {20 str:padright}   # pad each field in the group
                    map 
                        str:join .      # join the group into a string and print it
                }
                each
    console:reset
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



