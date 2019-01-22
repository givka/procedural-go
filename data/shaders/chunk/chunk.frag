#version 410 core

in vec3 Normal;
in vec3 FragPos;
in vec3 LightPos;
in vec4 MatColor;
in vec2 TexCoord;
in float Height;

out vec4 color;

uniform vec3 lightColor;
uniform sampler2D currentTexture;
uniform float near;
uniform int textureId;
uniform float far;

uniform sampler2D snowTexture; //0
uniform sampler2D rockTexture; //1
uniform sampler2D dirtTexture; //2
uniform sampler2D grassTexture;//3
uniform sampler2D sandTexture; //4

void setTextureCoefficients(inout float coeffs[5])
{
    if(Height > 1.4 && Normal.y < -0.9){
        coeffs[0] = 1.0;
        return;
    }
    if(Height > -0.5 && Normal.y < -0.9){
            coeffs[3] = 1.0;
            return;
    }
    if(Height > 1.5)
        coeffs[1] = 1.0;
    else if(Height > -0.5)
        coeffs[2] = 1.0;
/*    else if(Height > -0.1)
        coeffs[3] = 1.0;
*/    else if(Height > -2.0)
        coeffs[4] = 1.0;
    else coeffs[0] = 0.1;
}

float LinearizeDepth(float depth)
{
    float z = depth * 2.0 - 1.0; // back to NDC
    return (2.0 * near * far) / (far + near - z * (far - near));
}

void main()
{
    vec4 computedColor;
    float coeffs[5] = float[5](0.0, 0.0, 0.0, 0.0, 0.0);
    setTextureCoefficients(coeffs);
    if(textureId != 0){
    	computedColor =
    	coeffs[0] * texture(snowTexture, TexCoord)
    	+ coeffs[1] * texture(rockTexture, TexCoord)
    	+ coeffs[2] * texture(dirtTexture, TexCoord)
    	+ coeffs[3] * texture(grassTexture, TexCoord)
    	+ coeffs[4] * texture(sandTexture, TexCoord);
    	if(computedColor.a < 0.1)
    		discard;
    } else{
    		computedColor = MatColor;
    }


	// affects diffuse and specular lighting
	float lightPower = 4.0f;
	float ambientStrength = 0.3f;

	// diffuse and specular intensity are affected by the amount of light they get based on how
	// far they are from a light source (inverse square of distance)
	float distToLight = length(LightPos - FragPos);
	// this is not the correct equation for light decay but it is close
	// see light-casters sample for the proper way
	float distIntensityDecay = 1.0f / pow(distToLight, 2);

	vec3 ambientLight = ambientStrength * lightColor;

	vec3 norm = normalize(Normal);
	vec3 dirToLight = normalize(LightPos - FragPos);
	float lightNormalDiff = max(dot(norm, dirToLight), 0.0);

	// diffuse light is greatest when surface is perpendicular to light (dot product)
	vec3 diffuse = lightNormalDiff * lightColor;
	vec3 diffuseLight = lightPower * diffuse */* distIntensityDecay **/ lightColor;

	float specularStrength = 10.0f;
	int shininess = 64;
	vec3 viewPos = vec3(0.0f, 0.0f, 0.0f);
	vec3 dirToView = normalize(viewPos - FragPos);
	vec3 reflectDir = reflect(-dirToLight, norm);
	float spec = pow(max(dot(dirToView, reflectDir), 0.0), shininess);
	vec3 specularLight = lightPower * specularStrength * spec * distIntensityDecay * lightColor;

	vec3 result = (diffuseLight + specularLight + ambientLight) * computedColor.xyz;

    color = vec4(result, 1.0f);
	float depth = LinearizeDepth(gl_FragCoord.z) / far; // divide by far for demonstration
    color = mix(color, vec4(vec3(depth), 1.0), 0.5);
    //norm =

}
