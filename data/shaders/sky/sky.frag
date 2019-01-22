#version 410 core

in vec3 pos;

uniform sampler2D u_texture_tint;
uniform vec3 u_sun_pos;

out vec4 color;

void main()
{
    vec3 sun_norm = normalize(u_sun_pos);
    vec3 pos_norm = normalize(pos);

    float sun_height = -(sun_norm.y - 1.0) / 2.0;
    float distance_to_sun = length(sun_norm - pos_norm) / 2.0;
    float sun_size = 1.0 / 32.0;

    // sky
    color = texture(u_texture_tint, vec2(sun_height, 1.0 - distance_to_sun));

    // sun + halo
    color = mix(color, vec4(1.0, 1.0, 1.0, 1.0), sun_size / distance_to_sun);
}