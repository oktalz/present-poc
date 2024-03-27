var page = /^#?\d+$/.test(window.location.hash) ? parseInt(window.location.hash.slice(1), 10) : 0;
var spinner = false

function setSpinner(value){
    spinner = value
    if (value) {
        document.getElementById("spinner").classList.add("spinner")
    } else {
        document.getElementById("spinner").classList.remove("spinner")
    }
    
    document.querySelectorAll('.run').forEach(function(d) {
      if (spinner) {
        d.classList.add("closed")
      } else {
        d.classList.remove("closed")
      }
    })
    
}

function getSlideElements() {
  return document.querySelectorAll('[id^="slide-"]');
}

function updateSlideVisibility(page) {
  getSlideElements().forEach(function(slide) {
    if (parseInt(slide.id.slice(6)) == page) {
      slide.classList.remove('hidden');
    } else {
      slide.classList.add('hidden');
    }
  });
}

document.addEventListener("DOMContentLoaded", function() {
    updateSlideVisibility(page);
    document.querySelectorAll('pre code').forEach(function(codeElement) {
      codeElement.contentEditable = "true";
      codeElement.spellcheck = false;
    });
    document.querySelectorAll('code').forEach(function(codeElement) {
      if (codeElement.classList.length == 0) {
          codeElement.classList.add("hljs")
          codeElement.style.color = "#00ADD8"
          codeElement.style.backgroundColor = "#f8f8f8";
      }      
    });
});


document.addEventListener('keydown', function(e) {
  var keyCode = e.key;
  activeElement = document.activeElement;
  if (activeElement != null) {    
    if (keyCode == 'Escape' ) { 
      activeElement.blur();
    }
    if (activeElement.classList.contains('hljs') && (activeElement instanceof HTMLElement && activeElement.isContentEditable)) {
        return;
    }
  }
  if (keyCode == 'ArrowRight' || keyCode == 'ArrowDown' || keyCode == 'PageDown' || keyCode == ' ' ) {
    page = (page + 1);
    window.location.hash = page.toString();
    updateSlideVisibility(page);
  }
  if (keyCode == 'ArrowLeft' || keyCode == 'ArrowUp' || keyCode == 'PageUp') {
    page = (page - 1);
    if (page < 0) {
      page = 0;
    }
    window.location.hash = page.toString();
    updateSlideVisibility(page);
  }
  if (keyCode == 'r' ) {
    castTerminal()
  }
  if (keyCode == 'c' ) {
    const terminalElement = document.getElementById('terminal-'+page);
    terminalElement.innerHTML = ''
    terminalElement.classList.add('closed');
  }
});

window.addEventListener('wheel', function(event) {
  activeElement = document.activeElement;
  if (activeElement != null) {
    if (activeElement.classList.contains('hljs') && (activeElement instanceof HTMLElement && activeElement.isContentEditable)) {
        return;
    }
  }
  const eventObj = event || window.event; // cross browser
  var target = eventObj.target
  var rect = target.getBoundingClientRect()
  x = eventObj.clientX - rect.left,
  y = eventObj.clientY - rect.top;
  var elementUnderMouse = document.elementFromPoint(x, y); 
  if (elementUnderMouse != null) {
    if (elementUnderMouse.classList.contains('hljs') && (elementUnderMouse instanceof HTMLElement && elementUnderMouse.isContentEditable)) {
        return;
    }
  }

  const scrollDirection = event.deltaY > 0 ? 'downward' : 'upward';
  //console.log(`Mouse scroll ${scrollDirection}: ${Math.abs(event.deltaY)} pixels`);
  if (scrollDirection == 'downward') {
      page = (page + 1);
      window.location.hash = page.toString();
      updateSlideVisibility(page);
  }
  if (scrollDirection == 'upward') {
      page = (page - 1);
      if (page < 0) {
          page = 0;
      }
      window.location.hash = page.toString();
      updateSlideVisibility(page);
  }
}, true);

document.addEventListener('click', function(event) {
  var targetId = event.target.id;
  if (targetId && targetId.startsWith('run-')) {
    var pageNumber = targetId.substring(4); // strip 'run-' prefix
    console.log(`Clicked run button for page ${pageNumber}`);
    castTerminal()
  }
});
