#version 410 core

in vec3 Normal;
in vec3 FragPos;
in vec3 LightPos;
in vec4 MatColor;
in vec2 TexCoord;

out vec4 color;

uniform vec3 lightColor;
uniform sampler2D currentTexture;
uniform int textureId;
uniform float near; 
uniform float far; 
  
float LinearizeDepth(float depth) 
{
    float z = depth * 2.0 - 1.0; // back to NDC 
    return (2.0 * near * far) / (far + near - z * (far - near));	
}

void main()
{
    float lightPower = 5.0f;
    if(textureId == 0){
        lightPower = 2.0f;
    }
    float ambientStrength = 0.3f;
    float specularStrength = 20.0f;
	int shininess = 64;



	
	float distToLight = length(LightPos - FragPos);
	float distIntensityDecay = 1.0f / pow(distToLight, 2);
	vec3 ambientLight = ambientStrength * lightColor;
	vec3 norm = normalize(Normal);
	vec3 dirToLight = normalize(LightPos - FragPos);
	float lightNormalDiff = max(dot(norm, dirToLight), 0.0);
	vec3 diffuse = lightNormalDiff * lightColor;
	vec3 diffuseLight = lightPower * diffuse * lightColor;
	vec3 viewPos = vec3(0.0f, 0.0f, 0.0f);
	vec3 dirToView = normalize(viewPos - FragPos);
	vec3 reflectDir = reflect(-dirToLight, norm);
	float spec = pow(max(dot(dirToView, reflectDir), 0.0), shininess);
	vec3 specularLight = lightPower * specularStrength * spec * distIntensityDecay * lightColor;
	vec3 result = (diffuseLight + specularLight + ambientLight) * MatColor.xyz;

	if(textureId != 0){
		vec4 texColor = texture(currentTexture, TexCoord);
        texColor = mix(texColor, MatColor, 0.25);
		if(texColor.a < 0.1)
					discard;
        result = (diffuseLight + specularLight + ambientLight) * texColor.xyz;
		color = mix(texColor, vec4(result, 1.0f), 0.5);
	} else{
		color = vec4(result, 1.0f);
	}

	float depth = LinearizeDepth(gl_FragCoord.z) / far;
    color = mix(color, vec4(vec3(depth), 1.0), 0.5);
	
}
