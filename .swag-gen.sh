swag init --generalInfo cmd/horizon/main.go --output rest/docs && \
rm rest/docs/docs.go && \
rm rest/docs/swagger.yaml && \
mv rest/docs/swagger.json rest/docs/assets/swagger.json