document.addEventListener("DOMContentLoaded", function() {
  setPage(page);
  document.querySelectorAll('pre code').forEach(function(codeElement) {
    codeElement.contentEditable = "true";
    codeElement.spellcheck = false;
  });
  document.querySelectorAll('code').forEach(function(codeElement) {
    if (codeElement.classList.length == 0) {
        codeElement.classList.add("hljs")
        //codeElement.style.color = "#00ADD8" go colors
        codeElement.style.color = "#800"
        codeElement.style.backgroundColor = "#f8f8f8";
    }      
  });
  startWebsocket();
});
