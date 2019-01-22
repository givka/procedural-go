#version 410 core

in vec3 pos_sky;
in vec3 pos_stars;

uniform sampler2D u_texture_tint;
uniform vec3 u_sun_pos;
uniform mat4 u_rot_stars;

out vec4 color;

vec4 color_white = vec4(1.0, 1.0, 1.0, 1.0);

float random (vec3 st) {
    return fract(sin(dot(st,vec3(12.9898,78.233, 32.12324)))*43758.5453123);
}

void main()
{
    vec3 sun_norm = normalize(u_sun_pos);
    vec3 pos_norm = normalize(pos_sky);
    vec3 stars_norm = normalize(pos_stars);

    float sun_height = -(sun_norm.y - 1.0) / 2.0;
    float distance_to_sun = length(sun_norm - pos_norm) / 2.0;
    float sun_size = 1.0 / 32.0;

    // TODO: remove flickering
    // sky
    color = texture(u_texture_tint, vec2(sun_height, 1.0 - distance_to_sun));

    // sun
    color = mix(color, color_white, sun_size / distance_to_sun);

    // halo
    color = mix(color, color_white, (1.0 - distance_to_sun)/5.0);


    // stars
    float threshold = random(floor(stars_norm * 500.0));
    if (threshold > 0.995) {
      color = mix(color , color_white, 1.0 - sun_height);
    }

    // moon
    if(distance_to_sun > 0.9995){
      color = mix(vec4(0.5, 0.5, 0.5, 1.0), vec4(1.0, 1.0, 0.8, 1.0), sun_norm.y);
    }
}