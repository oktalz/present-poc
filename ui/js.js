var page = /^#?\d+$/.test(window.location.hash) ? parseInt(window.location.hash.slice(1), 10) : 0;
console.log("page is"+page)
setPage(page);
var spinner = false
var myID = ""
var showMenu = false

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

function setPage(newPage) { 
  if (newPage < -2){
    return
  }
  page = newPage 
  if (page < 0) {
    page = 0;
  }
  if (page > maxPage) {
    page = maxPage
  }
  window.location.hash = page.toString();
  updateSlideVisibility(page);
  menu = document.getElementById('menu')
  if (menu != null) {
    menu.classList.add('menu-hidden')
    showMenu = false
  }
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
    page = page + 1;
    setPage(page);
    updateData({
      Author: myID,
      Slide: page
    })
  }
  if (keyCode == 'ArrowLeft' || keyCode == 'ArrowUp' || keyCode == 'PageUp') {
    page = page - 1;
    setPage(page);  
    updateData({
      Author: myID,
      Slide: page
    })
  }
  if (keyCode == 'r' ) {
    castTerminal()
  }
  if (keyCode == 'm' ) {
    showMenu = !showMenu
    if (showMenu) {
      document.getElementById('menu').classList.remove('menu-hidden');
    } else {
      document.getElementById('menu').classList.add('menu-hidden');
    }
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
  const hoveredElements = document.querySelectorAll(':hover');
  const menuElements = Array.from(hoveredElements).filter((el) => el.classList.contains('menu'));
  if (menuElements.length > 0) {
    return;
  }

  const scrollDirection = event.deltaY > 0 ? 'downward' : 'upward';
  //console.log(`Mouse scroll ${scrollDirection}: ${Math.abs(event.deltaY)} pixels`);
  if (scrollDirection == 'downward') {
      page = page + 1;
      setPage(page);
      updateData({
        Author: myID,
        Slide: page
      })
  }
  if (scrollDirection == 'upward') {
      page = page - 1;
      setPage(page);
      updateData({
        Author: myID,
        Slide: page
      })
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

function tabChangeGlobal(tabID){
  console.log(tabID)
  let pageDIV = document.getElementById("slide-"+page);
  let tablinks = Array.from(pageDIV.querySelectorAll(".tablinks"));
  console.log(tablinks)
  for (let i = 0; i < tablinks.length; i++) {
    if (tablinks[i] && tablinks[i].getAttribute('id') === "tab-"+tabID) {
      tablinks[i].classList.add("active");
    } else {          
      tablinks[i].classList.remove("active");
    }
  }
  tablinks = Array.from(pageDIV.querySelectorAll(".tabcontent"));
  console.log(tablinks)
  for (let i = 0; i < tablinks.length; i++) {
    if (tablinks[i] && tablinks[i].getAttribute('id') === tabID) {
      tablinks[i].classList.remove("hidden-tab");          
    } else {
      tablinks[i].classList.add("hidden-tab");
    }       
  }
}
