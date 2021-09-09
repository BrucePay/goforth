###################################################################
#
# Mandebrot set rendered in character graphics by a GoForth script
#
###################################################################

40.0  -> screenY
120.0 -> screenX

-2.0  -> minReal
 1.0  -> maxReal
-1.2  -> minImaginary
 1.2  -> MaxImaginary

maxReal minReal - screenX  1.0 -  / -> realFactor
MaxImaginary minImaginary - screenY  1.0 - / -> imaginaryFactor

0.0  -> cImaginary
0.0  -> cReal

0.0  -> zReal
0.0  -> zImaginary

0.0  -> zRealSq
0.0  -> zImaginarySq

0    -> count
0    -> xOrd
0    -> yOrd
16   -> bailout 


# Initialize the color map, current and previous color variables
[
  ansi:black
  ansi:purple
  ansi:blue
  ansi:red
  ansi:cyan
  ansi:green
  ansi:yellow
  ansi:white
] -> color_map

ansi:black    -> lastcolor
color_map 0 @ -> color

cls

# Iterate vertically
{yOrd screenY 2 / < }
{
    MaxImaginary yOrd imaginaryFactor * - -> cImaginary
    
    # Iterate horizontally
    0 -> xOrd
    { xOrd screenX < }
    {
        xOrd realFactor * minReal + -> cReal
        cReal -> zReal
        cImaginary -> zImaginary
        
        0 -> count
        true -> continue    # bail out flag
        {count bailout < continue and}
        {
            zReal dup * -> zRealSq
            zImaginary dup * -> zImaginarySq
            zRealSq zImaginarySq + 4 >
            {
                false -> continue
            }
            {
                2.0 zReal * zImaginary * cImaginary + -> zImaginary 
                zRealSq zImaginarySq - cReal + -> zReal
            }
            ifte
            count succ -> count
        }
        while
        
        count bailout <
        {
            # Select the next color to use
            color_map count color_map len % @ -> color
            lastcolor color !=
            {
                color print
                color -> lastcolor
            }
            if

            # Render the two points
            xOrd int! yOrd int!                "#" console:print
            xOrd int! screenY yOrd - pred int! "#" console:print
        }
        if

        xOrd succ -> xOrd
    }
    while

    yOrd succ -> yOrd
}
while

0 41 console:at
console:reset

