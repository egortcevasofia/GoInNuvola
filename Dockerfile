FROM ubuntu:20.04
RUN apt-get update && apt-get install --yes nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

