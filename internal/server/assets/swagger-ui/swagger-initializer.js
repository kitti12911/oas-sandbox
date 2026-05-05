const script = document.currentScript;
const url = script && script.dataset.url ? script.dataset.url : "/openapi.json";

const methodOrder = {
    get: 0,
    post: 1,
    put: 2,
    patch: 3,
    delete: 4,
    head: 5,
    options: 6,
    trace: 7
};

const getValue = (operation, key) =>
    operation && typeof operation.get === "function"
        ? operation.get(key)
        : operation && operation[key];

const versionedPath = (path) => {
    const match = path.match(/^\/v(\d+)(\/.*)?$/);
    if (!match) {
        return {
            normalizedPath: path,
            version: 0
        };
    }

    return {
        normalizedPath: `/v{}${match[2] || ""}`,
        version: Number.parseInt(match[1], 10)
    };
};

const compareText = (left, right) => {
    if (left < right) return -1;
    if (left > right) return 1;
    return 0;
};

const compareOperations = (leftOperation, rightOperation) => {
    const leftMethod = String(
        getValue(leftOperation, "method") || ""
    ).toLowerCase();
    const rightMethod = String(
        getValue(rightOperation, "method") || ""
    ).toLowerCase();
    const methodCompare =
        (methodOrder[leftMethod] ?? Number.MAX_SAFE_INTEGER) -
        (methodOrder[rightMethod] ?? Number.MAX_SAFE_INTEGER);
    if (methodCompare !== 0) return methodCompare;

    const leftPath = String(getValue(leftOperation, "path") || "");
    const rightPath = String(getValue(rightOperation, "path") || "");
    const leftVersionedPath = versionedPath(leftPath);
    const rightVersionedPath = versionedPath(rightPath);
    const pathCompare = compareText(
        leftVersionedPath.normalizedPath,
        rightVersionedPath.normalizedPath
    );
    if (pathCompare !== 0) return pathCompare;

    const versionCompare =
        rightVersionedPath.version - leftVersionedPath.version;
    if (versionCompare !== 0) return versionCompare;

    return compareText(leftPath, rightPath);
};

window.onload = () => {
    window.ui = SwaggerUIBundle({
        url: url,
        dom_id: "#swagger-ui",
        deepLinking: true,
        operationsSorter: compareOperations,
        persistAuthorization: true,
        validatorUrl: null
    });
};
