#version 410

in vec3 vert; // Input center position for this hexagon
in float color; // Input color for this hexagon
in vec3 weights; // Input weights for this hexagon

out float vColor; // Color to be forwarded to geometry shader
out vec3 vWeights;

void main() {
	gl_Position = vec4(vert, 1);
	vColor = color;
	vWeights = weights;
}
