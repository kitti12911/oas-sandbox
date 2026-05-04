const script = document.currentScript;
const url = script && script.dataset.url ? script.dataset.url : "/openapi.json";

window.onload = () => {
  window.ui = SwaggerUIBundle({
    url: url,
    dom_id: "#swagger-ui",
    deepLinking: true,
    persistAuthorization: true,
    validatorUrl: null,
  });
};
