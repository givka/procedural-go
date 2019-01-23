#version 410 core

in vec3 Normal;
in vec3 FragPos;
in vec3 LightPos;
in vec4 MatColor;
in vec2 TexCoord;
in float Height;
in float RiverHeight;

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


const float minHeightSnow = 1.4;
const float maxNormalSnow = -0.9;
const float maxNormalSnowRock = -0.85;

const float minHeightSnowGrass = 1.3;

const float minRock = 1.5;

const float minRockDirt = 1.0;
const float minDirt = -0.5;

const float maxNormalGrass = -0.9;
const float maxNormalGrassDirt = -0.85;

const float minHeightGrass = -0.5;
const float minDirtSand = -1.5;
const float minSand = -2.0;

const float minWater =-2000.0;

//attempt to remove branching
void setTextureCoefficientsNoBranching(inout float coeffs[5])
{
    bool found = false;

    bool snowGrass = Height > minHeightSnowGrass && Height < minHeightSnow && Normal.y < maxNormalSnow || Normal.y < maxNormalGrass;
    bool snow =      !found && Height > minHeightSnow && Normal.y < maxNormalSnow;
    found = found || snow;
    bool snowRock =  !found && Height > minHeightSnow && Normal.y < maxNormalSnowRock;
    found = found || snowRock;
    bool grass =     !found && Height > minHeightGrass && Normal.y < maxNormalGrass;
    found = found || grass;
    bool grassDirt = !found && Height > minHeightGrass && Normal.y < maxNormalGrassDirt;
    found = found || grassDirt;

    bool rock =      !found && Height > minRock;
    found = found || rock;
    bool rockDirt =  !found && Height > minRockDirt;
    found = found || rockDirt;
    bool dirt =      !found && Height > minDirt;
    found = found || dirt;
    bool dirtSand =  !found && Height > minDirtSand;
    found = found || dirtSand;
    bool sand =      !found && Height > minSand;
    found = found || sand;
    bool def = !found;


    float a_height_snowgrass = smoothstep(minHeightSnowGrass, minHeightSnow, Height);
    float a_norm_snowrock = smoothstep(maxNormalSnowRock, maxNormalSnow, Normal.y);
    float a_norm_grassdirt = smoothstep(maxNormalGrassDirt, maxNormalGrass, Normal.y);

    float a_height_rockdirt = smoothstep(minRockDirt, minRock, Height);
    float a_height_dirtsand = smoothstep(minDirtSand, minDirt, Height);

    coeffs[0] = 1.0 * int(snow)   + a_height_snowgrass * int(snowGrass)      + a_norm_snowrock * int(snowRock);
    coeffs[1] = 1.0 * int(rock)   + a_height_rockdirt * int(rockDirt)        +(1 - a_norm_snowrock) * int(snowRock);
    coeffs[2] = 1.0 * int(dirt)   + (1 - a_height_rockdirt) * int(rockDirt)  +(1 - a_norm_grassdirt)* int(grassDirt) + a_height_dirtsand * int(dirtSand);
    coeffs[3] = 1.0 * int(grass)  + (1 - a_height_snowgrass) * int(snowGrass)+ a_norm_grassdirt * int(grassDirt);
    coeffs[4] = 1.0 * int(sand)   + (1 - a_height_dirtsand) * int(dirtSand);
}

void setTextureCoefficients(inout float coeffs[5])
{
    if(Height > minHeightSnowGrass && Height < minHeightSnow && Normal.y < maxNormalSnow || Normal.y < maxNormalGrass){
            float a_height = smoothstep(minHeightSnowGrass, minHeightSnow, Height);
            coeffs[0] = a_height;
            coeffs[3] = 1 - a_height;
            return;
    }
    if(Height > minHeightSnow){
        if(Normal.y < maxNormalSnow){
            coeffs[0] = 1.0;
            return;
        }else if(Normal.y < maxNormalSnowRock){
            float a = smoothstep(maxNormalSnowRock, maxNormalSnow, Normal.y);
            coeffs[0] = a;
            coeffs[1] = 1 - a;
            return;
        }
    }
    if(Height > minHeightGrass){
        if(Normal.y < maxNormalGrass){
            coeffs[3] = 1.0;
            return;
        }else if(Normal.y < maxNormalGrassDirt){
            float a = smoothstep(maxNormalGrassDirt, maxNormalGrass, Normal.y);
            coeffs[3] = a;
            coeffs[2] = 1 - a;
            return;
        }
    }
    if(Height > minRock)
        coeffs[1] = 1.0;
    else if (Height > minRockDirt){
        float a = smoothstep(minRockDirt, minRock, Height);
        coeffs[1] = a;
        coeffs[2] = 1 - a;
    }
    else if(Height > minDirt)
        coeffs[2] = 1.0;
    else if (Height > minDirtSand){
        float a = smoothstep(minDirtSand, minDirt, Height);
        coeffs[2] = a;
        coeffs[4] = 1 - a;
    }
    else if(Height > minSand)
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
    //setTextureCoefficientsNoBranching(coeffs);
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
    if(RiverHeight > 2.5 /*&& Height < minRock*/ || Height < minSand + 0.1)
    {
        color = vec4(0.0, 0.0, 1.0, 1.0);
    }
    //color.b = 1 - smoothstep(-3.0, -1.0, RiverHeight);
    //norm =

}
