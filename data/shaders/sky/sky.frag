#version 410 core

in vec3 pos;

uniform float hours;
uniform float radius;
uniform sampler2D currentTexture;
uniform vec3 sun_pos;


out vec4 color;




void main()
{
    float tint = hours / 12.0;

    vec3 sun_norm = normalize(sun_pos);
    vec3 pos_norm = normalize(pos);

    float altitude = pos_norm.y;
    altitude = altitude;



    float dist = length(sun_norm - pos_norm);

    color = texture(currentTexture, vec2(tint, altitude));

    if(dist < 0.1){
        color = vec4(1,1,1,1);
    }
    

}