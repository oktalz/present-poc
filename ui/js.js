var page = /^#?\d+$/.test(window.location.hash) ? parseInt(window.location.hash.slice(1), 10) : 0;
console.log("page is "+page)
setPage(page);
var spinner = false
var myID = ""
var showMenu = false

function setSpinner(value){
    spinner = value
    if (value) {
        document.getElementById("run-"+page+"-refresh").classList.remove("closed")
        document.getElementById("run-"+page+"").classList.add("closed")
    } else {
        document.getElementById("run-"+page+"-refresh").classList.add("closed")
        document.getElementById("run-"+page+"").classList.remove("closed")
    }
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
  if (keyCode == 'ArrowRight' || keyCode == 'ArrowUp' || keyCode == 'PageDown' || keyCode == ' ' ) {
    oldPage = page;
    page = page + 1;
    target = getPageUp(oldPage,page);
    setPage(target);
    updateData({
      Author: myID,
      Slide: page
    })
  }
  if (keyCode == 'ArrowLeft' || keyCode == 'ArrowDown' || keyCode == 'PageUp') {
    oldPage = page;    
    page = page - 1;
    target = getPageDown(oldPage,page);
    setPage(target);  
    updateData({
      Author: myID,
      Slide: page
    })
  }
  if (keyCode == 'r' || keyCode == 'e' ) {
    castTerminal()
  }
  if (keyCode == 'c' || keyCode == 'b' ) {
    const terminalElement = document.getElementById('terminal-'+page);
    terminalElement.innerHTML = ''
    terminalElement.classList.add('closed');
  }
  if (keyCode == 'm' ) {
    showMenu = !showMenu
    if (showMenu) {
      document.getElementById('menu').classList.remove('menu-hidden');
    } else {
      document.getElementById('menu').classList.add('menu-hidden');
    }
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
  menuElements = Array.from(hoveredElements).filter((el) => el.classList.contains('menu'));
  if (menuElements.length > 0) {
    return;
  }
  menuElements = Array.from(hoveredElements).filter((el) => el.classList.contains('box-overflow'));
  if (menuElements.length > 0) {
    return;
  }

  const scrollDirection = event.deltaY > 0 ? 'downward' : 'upward';
  //console.log(`Mouse scroll ${scrollDirection}: ${Math.abs(event.deltaY)} pixels`);
  if (scrollDirection == 'downward') {
      oldPage = page;
      page = page + 1;
      target = getPageUp(oldPage,page);
      setPage(target);
      updateData({
        Author: myID,
        Slide: page
      })
  }
  if (scrollDirection == 'upward') {
      oldPage = page;
      page = page - 1;
      target = getPageDown(oldPage,page);
      setPage(target);
      updateData({
        Author: myID,
        Slide: page
      })
  }
}, true);

document.addEventListener('click', function(event) {
  var targetId = event.target.id;
  if (targetId && targetId.startsWith('run-icon-')) {
    var pageNumber = targetId.substring(9); // strip 'run-' prefix
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

function correctD2Graph(svg){
  svg.setAttribute('width', "100%");
  svg.setAttribute('height', "100%");
  svg.parentElement.setAttribute('width', "100%");
  svg.parentElement.setAttribute('height', "100%");
  svg.parentElement.setAttribute('preserveAspectRatio', "YMin meet");
}
