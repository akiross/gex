#version 410 core
layout (points) in;
layout (triangle_strip, max_vertices = 22) out;
in float vColor[]; // Color of each input vertex (just 1)
in vec3 vWeights[]; // Weight for each input vertex (just 1)

out float gColor; // Color for output primitives

const float PI = 3.14159265;
const float SD = HEX_SIDE * 0.5 / PHO; // Min distance between center and hex sides

const float radius = 0.75; // Relative radius for hex (1.0 is full tessellation)
const float foglia = 0.5; // Relative size for weight rectangle (1.0 is filling space between hexes)

const vec4 scale = vec4(INV_ASPECT_RATIO, 1.0, 1.0, 1.0);

/*
Il raggio è la distanza AI VERTICI, noi vogliamo la distanza ALLA FACCIA
*/

// Compute position of i-th vertex using sine and cosine
vec2 vert(int i, float ra) {
	float a = PI * (0.5 + i / 3.0);
	return vec2(ra * cos(a), ra * sin(a));
}

// Emits position of ith vertex
void pos(int i) {
	gl_Position = scale * (gl_in[0].gl_Position + vec4(vert(i, radius * SD), 0.0, 0.0));
	EmitVertex();
}

// Has to emit 2 vertices for the specified side
void opos(int i, int j, float o) {
	vec2 v1 = vert(i, radius * SD);
	vec2 v2 = vert(j, radius * SD);
	// Offset direction
	float a = PI * 2.0/3.0 + PI * i / 3.0;
	vec2 dir = vec2(cos(a), sin(a));

	float k = HEX_SIDE - 2.0 * radius * SD * PHO;

	gl_Position = scale * (gl_in[0].gl_Position + vec4(v1 + dir * k * o, 0.0, 0.0));
	EmitVertex();
	gl_Position = scale * (gl_in[0].gl_Position + vec4(v2 + dir * k * o, 0.0, 0.0));
	EmitVertex();
}

// Produce two triangle strips using vertices (top vertex is number 0, CCW)
void main() {
	if (true) {
		gColor = vColor[0];
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

	float w;
	if (true) {
		w = 1.0 - vWeights[0].x;
		gColor = 1.0 - w;
		opos(0, 1, w * 0.5);
		opos(0, 1, 1.0 - w * 0.5);
		EndPrimitive();

		w = 1.0 - vWeights[0].y;
		gColor = 1.0 - w;
		opos(5, 0, w * 0.5);
		opos(5, 0, 1.0 - w * 0.5);
		EndPrimitive();

		w = 1.0 - vWeights[0].z;
		gColor = 1.0 - w;
		opos(4, 5, w * 0.5);
		opos(4, 5, 1.0 - w * 0.5);
		EndPrimitive();
	}
}
