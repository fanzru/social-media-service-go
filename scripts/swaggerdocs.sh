#!/bin/bash

########
# Help #
########
Help() {
        # Display Help
        echo "Convert yaml inside api/http from openapi 3 to swagger 2."
        echo
        echo "Usage:"
        echo "  ./swaggerdocs.sh [options]"
        echo
        echo "options:"
        echo "h    Display Help"
        echo "b    Change base file (default: empty)"
        echo
}

# Set error handling
set -e

files=""

while getopts ":hb:" flag;
do
        case "$flag" in
                h) Help
                   exit;;
                b) files="api/doc/swagger/$OPTARG.json";;
                \?) echo "Illegal option(s)"
                    exit 1;;
        esac
done

# Check if api/http directory exists
if [ ! -d "./api/http" ]; then
        echo "Error: ./api/http directory not found"
        exit 1
fi

# Check if there are any yaml files
yaml_count=$(ls ./api/http/*.yaml 2>/dev/null | wc -l)
if [ "$yaml_count" -eq 0 ]; then
        echo "Error: No yaml files found in ./api/http/"
        exit 1
fi

cnt=0

for file in ./api/http/*.yaml; do
        f=$(basename "$file" .yaml)
        mkdir -p ./api/doc/swagger

        echo "Converting $file to Swagger 2.0..."
        
        # Convert OpenAPI 3 to Swagger 2
        if ! api-spec-converter \
                --from=openapi_3 \
                --to=swagger_2 \
                "$file" > "api/doc/swagger/$f.json"; then
                echo "Error: Failed to convert $file"
                exit 1
        fi

        files="$files api/doc/swagger/$f.json"
        cnt=$((cnt+1))
done

# Create docs directory
mkdir -p ./docs/swagger

echo "Merging swagger files..."
# Merge all swagger files into one
if [ $cnt -eq 1 ]; then
        # Only one file, just copy it
        first_file=$(echo $files | awk '{print $1}')
        if ! cp "$first_file" docs/swagger/docs.json; then
                echo "Error: Failed to copy swagger file"
                exit 1
        fi
else
        # Multiple files - merge them manually using jq
        echo "Merging $cnt swagger files..."
        
        # Create a temporary merged file
        temp_merged=$(mktemp)
        
        # Start with the first file as base
        first_file=$(echo $files | awk '{print $1}')
        cp "$first_file" "$temp_merged"
        
        # Merge remaining files
        for file in $files; do
                if [ "$file" != "$first_file" ]; then
                        echo "Merging $file..."
                        # Use jq to merge paths and components with null safety
                        jq -s '
                                .[0] as $base |
                                .[1] as $merge |
                                $base | 
                                .paths = (($base.paths // {}) + ($merge.paths // {})) |
                                .components.schemas = (($base.components.schemas // {}) + ($merge.components.schemas // {})) |
                                .tags = (($base.tags // []) + ($merge.tags // []) | unique)
                        ' "$temp_merged" "$file" > "${temp_merged}.tmp" && mv "${temp_merged}.tmp" "$temp_merged"
                fi
        done
        
        # Copy merged result to final destination
        if ! cp "$temp_merged" docs/swagger/docs.json; then
                echo "Error: Failed to save merged swagger file"
                rm -f "$temp_merged"
                exit 1
        fi
        
        # Clean up temporary file
        rm -f "$temp_merged"
fi

echo "Swagger documentation generated successfully at docs/swagger/docs.json"
