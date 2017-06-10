#version 410 core
layout (points) in;
layout (triangle_strip, max_vertices = 22) out;
in float vColor[]; // Color of each input vertex (just 1)
out float gColor; // Color for output primitives

const float PI = 3.14159265;
const float r = 0.06;
const float rr = 0.025; // Ring distance
const float rs = 0.001; // Ring size

const vec4 scale = vec4(3.0/4.0, 1.0, 1.0, 1.0);

// Compute position of i-th vertex using sine and cosine
vec2 vert(int i, float ra) {
	float a = PI * (0.5 + i / 3.0);
	return vec2(ra * cos(a), ra * sin(a));
}

// Emits position of ith vertex
void pos(int i) {
	gl_Position = scale * (gl_in[0].gl_Position + vec4(vert(i, r), 0.0, 0.0));
	EmitVertex();
}

// Has to emit 2 vertices for the specified side
void opos(int i, int j, float o) {
	vec2 v1 = vert(i, r);
	vec2 v2 = vert(j, r);
	// Offset direction
	float a = PI * 2.0/3.0 + PI * i / 3.0;
	vec2 dir = vec2(cos(a), sin(a));

	gl_Position = scale * (gl_in[0].gl_Position + vec4(v1 + dir * o, 0.0, 0.0));
	EmitVertex();
	gl_Position = scale * (gl_in[0].gl_Position + vec4(v2 + dir * o, 0.0, 0.0));
	EmitVertex();
}

// Produce two triangle strips using vertices (top vertex is number 0, CCW)
void main() {
	gColor = vColor[0];

	if (true) {
		pos(1);
		pos(0);
		gl_Position = scale * gl_in[0].gl_Position;
		pos(5);
		pos(4);
		pos(1);
		pos(2);
		gl_Position = scale * gl_in[0].gl_Position;
		pos(3);
		pos(4);
		EndPrimitive();
	}

	if (true) {
		opos(0, 1, rr - rs);
		opos(0, 1, rr + rs);
		EndPrimitive();

		opos(5, 0, rr - rs);
		opos(5, 0, rr + rs);
		EndPrimitive();

		opos(4, 5, rr - rs);
		opos(4, 5, rr + rs);
		EndPrimitive();
	}
}
