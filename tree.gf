######################################################################################
#
# GoForth example showing how to define, build and print a tree
#
######################################################################################

DEFINE node:new ==
    [] dict! -> this
    $this swap :right swap ! 
    $this swap :left swap ! 
    $this swap :value swap !
    $this
;

DEFINE node:print ==
    dup
    {
        dup :left  @ node:print
        dup :right @ node:print
            :value @ .
    }
    {
        pop
    }
    ifte
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

    
