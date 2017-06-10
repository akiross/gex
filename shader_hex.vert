#version 410

in vec3 vert; // Input center position for this hexagon
in float color; // Input color for this hexagon

out float vColor; // Color to be forwarded to geometry shader

void main() {
	//gl_Position = vec4(3.0 / 4.0, 1.0, 1.0, 1.0) * vec4(vert, 1);
	gl_Position = vec4(vert, 1);
	vColor = color;
}
