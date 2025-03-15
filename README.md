# suggestions
Basic implementation of Angular webapp with a restful go backend


# docker
docker build -t suggestions:latest .
docker run -p 8080:8080 --volume ./data:/data suggestions:latest


# with provided compose
docker compose build
docker compose up