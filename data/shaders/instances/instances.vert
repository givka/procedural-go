#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec4 color;
layout (location = 3) in vec2 texture;
layout (location = 4) in mat4 transform;

uniform mat4 view;
uniform mat4 project;
uniform mat4 model;
uniform vec3 lightPos;
uniform int u_nbr_instances;
uniform int textureId;

out vec3 Normal;
out vec3 FragPos;
out vec3 LightPos;
out vec4 MatColor;
out vec2 TexCoord;

float random (vec2 st) {
    return fract(sin(dot(st,vec2(12.9898,78.233)))*43758.5453123);
}

void main()
{
    gl_Position = project * view * transform * model * vec4(position, 1.0);
    FragPos = position;
    LightPos = lightPos;
    Normal = (transpose(inverse(model)) * vec4(normal, 1.0)).xyz;
    TexCoord = texture; 
    MatColor = color;

    // random grass color
    if(textureId == 0){
        float nbr_instances = u_nbr_instances;
        float instance_id = gl_InstanceID;
        float threshold = random(vec2(instance_id/nbr_instances * 500.0));
        MatColor = mix(color, vec4(color.x, 0.25, color.z, 1.0), threshold);
    }

}
