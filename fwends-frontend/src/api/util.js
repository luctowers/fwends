export function lazyPromise(fn) {
  let promise;
  return function() {
    if (!promise) {
      promise = fn();
    }
    return promise;
  };
}

// derived from https://stackoverflow.com/questions/950087/how-do-i-include-a-javascript-file-in-another-javascript-file
export function dynamicScriptLoad(src, timeout) {
  return new Promise((resolve, reject) => {
    let head = document.head;
    let script = document.createElement("script");
    script.type = "text/javascript";
    script.src = src;
    script.onreadystatechange = resolve;
    script.onload = resolve;
    head.appendChild(script);
    setTimeout(() => {
      head.removeChild(script);
      reject(new Error("Dynamic script load timed out"));
    }, timeout || 10000);
  });
}
