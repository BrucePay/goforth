######################################################################################
#
# GoForth example showing how to define, build and print a tree
#
######################################################################################

DEFINE node:new value left right ==
    [
        :value value
        :left left
        :right right
    ]
    dict!
;

DEFINE node:print n ==
    n
    {
        n :left  @ node:print
        n :right @ node:print
        n :value @ .
    }
    if
;

# Create and print a tree
7
    5
        3
            1 nil nil node:new
            2 nil nil node:new
        node:new
        4 nil nil node:new
    node:new    
    6 nil nil node:new
        node:new

dup .yellow

node:print

    
