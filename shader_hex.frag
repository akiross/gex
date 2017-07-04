#version 410
in float gColor; // Color from geometry shader
out vec4 oColor; // Color of fragment
void main() { oColor = vec4(1.0 - gColor, 1.0 - gColor, 1.0 - gColor, 1.0); }
