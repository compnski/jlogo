# jlogo
Toy logo interpreter that can control a physical rasberry-pi based turtle.

# Code
Parsing is done with the https://github.com/alecthomas/participle library, which creates a grammar from your structs, like JSON parsing.
You get back the AST in the form of the structs you create.
I borrowed heavily from the BASIC example to get support for expressions.

# Videos
* Mark II in action: https://youtu.be/Cr2jLVUYcts
* Mark I in action : https://youtu.be/Abdd7Wg_pFM

# Building it
## Tinkercard Links:
* Turtle Base: https://www.tinkercad.com/dashboard?type=tinkercad&collection=projects&id=28NQXupXaBR
* Pen tube: *https://www.tinkercad.com/dashboard?type=tinkercad&collection=projects&id=28NQXupXaBR
* Wheel (But prefer rubber 65mm wheels): https://www.tinkercad.com/dashboard?type=tinkercad&collection=projects&id=28NQXupXaBR

## Parts
* 5v power regulator: https://www.amazon.com/gp/product/B076P4C42B
* SG90 Servo: https://www.amazon.com/gp/product/B07Q6JGWNV
* 65m Rubber wheel: https://www.amazon.com/gp/product/B07GGXFFSR
* 28BYJ Stepper + ULN2003 Driver: https://www.amazon.com/gp/product/B01CP18J4A
* Battery Charging Module: https://www.amazon.com/gp/product/B08QD3W677
* 3.7v Lipo battery: https://www.amazon.com/gp/product/B0867KDMY7
* Rasberry Pi
* Assorted wires, connectors, etc
  * JST 1.25mm connectors: https://www.amazon.com/gp/product/B013JRWCBU
  * JST 20AWG connectors: https://www.amazon.com/gp/product/B01M5AHF0Z

# TODO
* Better calibration for turns
* Progress report during long drawings
* Cancel action in progress
* Estimated runtime when drawing 
* LOGO procedures
* Additional language features
