<template>
  <div class="presenter-time" v-if="presenterMode">
    {{timer}}
  </div>
  <div class="presenter-comment" v-if="presenterMode">
    {{slides?.[state.page].notes}}
  </div>
  <div class="menu" v-if="showMenu">
    <table>
      <tr v-for="(slide, index) in slides" :key="index" >
        <template v-if="lastTitle(index)">
          <td @click="menuClick(index)" style="cursor: pointer;">{{ slide.page }}</td>
          <td>&nbsp;&nbsp;&nbsp;</td>
          <td @click="menuClick(index)" style="cursor: pointer;">{{ title(slide.markdown) }}</td>
        </template>
      </tr>
    </table>
  </div>
  <div @wheel="onWheel" class="slide" :id="'slide-' + index" :ref="'slide' + index"
      v-bind:class="{ 'hidden': index!==state.page, 'page-not-print-visible': slide.print_page < 1 }"
      v-for="(slide, index) in slides" :key="index"
      v-bind:style="{
        'font-size': slide.font_size,
        'background-image': 'url(' + slide.background + ')',
        'background-color': slide.background_color
      }"
  >
    <div class="slide-data">
      <div v-html="md.render(slide.markdown)"></div>
      <div
        class="terminal"
        v-html="state.terminal[index]"
        v-if="state.terminal[index] != ''">
      </div>
      <div
        class="loading_cast"
        v-if="state.terminal_loading"
        v-html="md.render(`{green}(:fa-spinner#fa-spin:)`)"
      >
        
      </div>

      <div class="page-num">
        <div
          v-if="RunVisible(state.page)"
          class="button run-button"
          @click="execTerm"
          >
            Run
        </div>
        <div
          v-if="state.terminal[index] == '-1'"
          class="button run-button"
          @click="state.terminal[index]=''"
          >
            Close
        </div>
        &nbsp;&nbsp;&nbsp;<span class="view-page">{{ slide.page }}</span><span class="view-print-page">{{ slide.print_page }}</span>
      </div>
      <div class="page-break"></div>
    </div><!-- <div class="slide" -->
  </div> <!-- <div class="slide" -->
</template>

<script lang="ts">
import { defineComponent, ref, reactive } from 'vue';
import MarkdownIt from 'markdown-it';
import emoji from 'markdown-it-emoji';
import highlightjs from 'markdown-it-highlightjs';

import graph from 'markdown-it-textual-uml';
import 'highlight.js/styles/default.css';

import markdownItColorInline from 'markdown-it-color-inline'
import { align } from "@mdit/plugin-align";
import { attrs } from "@mdit/plugin-attrs";
import { footnote } from "@mdit/plugin-footnote";
import { imgSize } from "@mdit/plugin-img-size";
import 'markdown-it-icons/dist/index.css'
import * as AsciinemaPlayer from 'asciinema-player';
import '@fortawesome/fontawesome-free/css/all.css'

import zbTabs from '../plugins/Tabs';
import zbStyle from '../plugins/Style';
import zbTable from '../plugins/Table';
import zbImage from '../plugins/Image';
import zbFontAwesome from '../plugins/FontAwesome';

export default defineComponent({
  props: {
    syncMessage: Object,
  },
  watch: {
    syncMessage(newValue, oldValue) {
      // Perform actions based on the prop change
      oldValue = oldValue
      this.state.myID = newValue.ID
      this.state.page = newValue.Slide
      this.setPage(false)
    }
  },
  components: {
  },
  methods: {
    onWheel: function(e: WheelEvent) {
      if (this.slides == null) {
        return
      }
      if (e.currentTarget == null) {
        return
      }
      // console.log("wheeling over " + (e.currentTarget as HTMLElement).id, e.deltaY);
      if (e.deltaY > 0) {
        if (this.state.page < this.slides.length-1) {
          this.state.page+= 1
          this.setPage(true)
        }
      }
      if (e.deltaY < 0) {
        if (this.state.page > 0) {
          this.state.page-= 1
          this.setPage(true)
        }
      }
      e.stopPropagation();
    },
    newTitle: function (index: number) {
      if (index == 0) {
        return true
      }
      if (this.slides == null) {
        return true
      }
      if (this.slides[index] == null) {
        return true
      }
      let curTitle = this.title(this.slides[index].markdown)
      let prevTitle = this.title(this.slides[index-1].markdown)
      return (curTitle != prevTitle)
    },
    lastTitle: function (index: number) {
      if (this.slides == null) {
        return true
      }
      if (index == this.slides.length-1) {
        return true
      }
      if (this.slides[index] == null) {
        return true
      }
      let curTitle = this.title(this.slides[index].markdown)
      let prevTitle = this.title(this.slides[index+1].markdown)
      return (curTitle != prevTitle)
    },
    title: function (textBlock: string) {
      let lines = textBlock.split('\n');
      for (let i = 0; i < lines.length; i++) {
        let line = lines[i];
        let hashIndex = line.lastIndexOf('#');
        if (hashIndex !== -1) {
          return line.slice(hashIndex + 1).trim();
        }
      }
      return "";
    },
    menuClick: function (index: number) {
      this.showMenu = false
      this.state.page = index
      this.setPage(true)
    },
    execTerm: function () {
      if (this.state.terminal_loading) {
        return
      }if (this.slides == null) {
        return
      }

      let activeElement = document.activeElement;
      if (activeElement == null) {
        return;
      }
      if (activeElement.classList.contains('code')) {
          return;
      }

      const slideElement = document.getElementById('slide-'+this.state.page);
      let codeText: string[] = [];
      if (slideElement) {
        let codeElements = slideElement.querySelectorAll('pre code');
        codeText = Array.from(codeElements).map(codeElement => (codeElement as HTMLElement).innerText);
      }

      //const baseUrl = import.meta.env.VITE_BASE_URL
      const baseUrlStart = import.meta.env.VITE_BASE_URL.replace(/^https?:\/\//, '');
      const baseUrl = baseUrlStart.replace(/^(\/|\\)/, '');
      console.log(baseUrl)
      //=====================================================================
      this.state.terminal[this.state.page] = ""
      this.state.terminal_loading = true
      //const evtSource = new EventSource(baseUrl+"/cast-sse?slide="+this.state.page);
      // Create WebSocket connection
      const socket = new WebSocket('ws://'+baseUrl+"/cast");

      // Connection opened
      socket.addEventListener('open', () => {
        //console.log(codeText)
        let body = JSON.stringify({slide: this.state.page, code: codeText})
        console.log(body)
        socket.send(body);
        this.CheckTabState(codeText)
      });

      // Listen for messages
      socket.addEventListener('message', (event) => {
          console.log('Message from server: ', event.data);
          if (this.state.terminal[this.state.page] != "") {
            this.state.terminal[this.state.page] += '<br>'
          } 
          this.state.terminal[this.state.page] += event.data
          this.CheckTabState(codeText)
      });
      
      socket.onclose = () => {
        console.log('Socket is closed');
        socket.close();
        this.state.terminal_loading = false        
        
      };
    },
    tabChange: function(tabID: string) {
      console.log(tabID)
      let tablinks = document.getElementsByClassName("tablinks");
      for (let i = 0; i < tablinks.length; i++) {
        if (tablinks[i] && tablinks[i].getAttribute('id') === "tab-"+tabID) {
          if (!tablinks[i].classList.contains("active")) {          
            tablinks[i].classList.add("active");
          } 
        } else {          
          tablinks[i].classList.remove("active");
        }
      }
      tablinks = document.getElementsByClassName("tabcontent");
      for (let i = 0; i < tablinks.length; i++) {
        if (tablinks[i] && tablinks[i].getAttribute('id') === tabID) {
          tablinks[i].classList.remove("hidden");          
        } else {
          if (!tablinks[i].classList.contains("hidden")) {          
            tablinks[i].classList.add("hidden");
          }
        }       
      }
    },
    CheckTabState: function(codeText: string[]) {
      // this code here is because I'm lazy to understand why tabs are reseting data and editable state
      // TODO: fix this 
      let codeSelector = document.getElementById('slide-' + this.state.page)
      if (codeSelector) {
        setTimeout(() => {
          if (codeSelector) {
            let codeElements = codeSelector.querySelectorAll('pre code')
            for (let i = 0; i < codeElements.length; i++) {
              let codeElement = codeElements[i] as HTMLElement
              codeElement.contentEditable = "true";
              codeElement.spellcheck = false;
              if (codeElement.innerText != codeText[i]){
                codeElement.innerText = codeText[i];
              }
            }
          }            
        }, 10);
      }
    },
    RunVisible: function (index: number) {
      if (index==null || this.slides == null) {
        return false
      }
      if (this.slides[index].terminal) {
        return true
      }
      if (this.slides[index].cast) {
        return true
      }
      if (this.slides[index].run) {
        return true
      }
      let pl = this.state.player[index]
      if (pl == null) {
        return false
      }
      if (pl.player != null) {
        return true
      }
      return false
    },
    handleKeyPress: function (e: KeyboardEvent) {
      const keyCode = e.key;

      let activeElement = document.activeElement;
      if (activeElement != null) {
        if (activeElement.classList.contains('hljs') && (activeElement instanceof HTMLElement && activeElement.isContentEditable)) {
            return;
        }
      }
      if (this.slides == null) {
        return
      }
      if (keyCode == 'ArrowRight' || keyCode == 'ArrowDown' || keyCode == 'PageDown') {
        if (this.state.page < this.slides.length-1) {
          this.state.page+= 1
          this.setPage(true)
        }
      }
      if (keyCode== 'ArrowLeft' || keyCode == 'ArrowUp' || keyCode == 'PageUp') {
        if (this.state.page > 0) {
          this.state.page-= 1
          this.setPage(true)
        }
      }
      if (keyCode == 'r') {
        this.execTerm()
      }
      if (keyCode == 'm') {
        this.showMenu = !this.showMenu
      }
      if (keyCode == 'c') {
        this.state.terminal[this.state.page] = '';        
      }
      if (keyCode == ' ') {
      }
      //console.log(this.state, keyCode, e, e.key);
    },
    setPage: function(isLocal: boolean) {
      if (this.slides == null){
        return
      }
      window.location.hash = (this.state.page+1).toString()
      if (this.slides[this.state.page].background == ""){
        document.body.style.backgroundImage = 'none'
      }else{
        document.body.style.backgroundImage = 'url('+ this.slides[this.state.page].background+')'
      }
      //document.body.style.fontSize = this.slides[this.state.page].font_size;
      if (this.slides[this.state.page].can_edit){
        let codeSelector = document.getElementById('slide-' + this.state.page)
        if (codeSelector) {
          setTimeout(() => {
            if (codeSelector) {
              let codeElements = codeSelector.querySelectorAll('pre code')
              for (let i = 0; i < codeElements.length; i++) {
                let codeElement = codeElements[i] as HTMLElement
                codeElement.contentEditable = "true";
                codeElement.spellcheck = false;
              }
            }            
          }, 10);
        }
      }
      const baseUrl = import.meta.env.VITE_BASE_URL
      if (this.state.myID == -1){
        return
      }
      if (!isLocal) {
        return
      }
      fetch(baseUrl + "/update", {
              method: 'POST',
              headers: {
                'Content-Type': 'text/plain'
              },
              body: JSON.stringify({
                Author: this.state.myID,
                Slide: this.state.page
              })
          })
          .then(response => response.text())
          // .then(data => {
          //   return processData(data)
          // })
      //console.log(this.slides[this.state.page])
    }
  },
  created() {

  },
  beforeDestroy() {
    // Remember to remove the event listener when the component is destroyed
  },
  mounted() {
    window.addEventListener('keydown', this.handleKeyPress)
    this.$nextTick(function () {
      this.setPage(true)
      setTimeout(() => {
        this.setPage(true)
        // @ts-ignore
        window.tabChangeGlobal = this.tabChange;
      }, 100);
    })
  },
  unmounted() {
    window.removeEventListener('keydown', this.handleKeyPress)
  },
  setup() {
    const state = reactive({
        page: 0,
        myID: -1,
        terminal: [] as string[],        
        terminal_loading: false,
        player: [] as any
      });

    const md = new MarkdownIt();
    md.use(emoji);
    md.use(highlightjs, {
      inline: true
    })
    md.use(graph);
    md.use(markdownItColorInline)
    md.use(align)
    md.use(attrs)
    md.use(footnote)
    md.use(imgSize)

    md.use(zbTabs)
    md.use(zbStyle)
    md.use(zbImage)
    md.use(zbTable)
    md.use(zbFontAwesome)

    //md.use(tasklist)

    type Slide = {
      background: string;
      background_color: string;
      font_size: string;
      terminal: boolean;
      markdown: string;
      notes: string;
      can_edit: boolean;
      cast: boolean;
      run: boolean;
      page: number;
      print_page: number;
    };

    //const slides = ref(null);
    const slides = ref<Slide[] | null>(null);
    const baseUrl = import.meta.env.VITE_BASE_URL
    fetch(baseUrl+'/api')
      .then(response => response.json())
      .then(data => {
        slides.value = data;
        return data;
      })
      .then(data => {
        if (state.terminal == null){
          return
        }
        // Process and log the data here
        for (let i = 0; i < data.length; i++) {
          state.terminal[i]=""
          state.player[i] = {player: null,visible: false}
          if (data != null && data[i].asciinema != null){
            const playerElement = document.getElementById('ascinema-player-'+i.toString());
            if (playerElement) {
              let mov = {}
              if (data[i].asciinema.url != ''){
                mov = data[i].asciinema.url
              } else {
                mov = {data: data[i].asciinema.cast}
              }

              let player = AsciinemaPlayer.create(
                mov,
                playerElement,
                {
                  //poster: "data:text/plain,I'm regular \x1b[1;32mI'm bold green\x1b[3BI'm 3 lines down",
                  fit: "height",
                  poster: 'npt:0:0'
                  //terminalFontSize: ""
                })
              player.play();
              player.pause();
              console.log(player)
              state.player[i] = {player: player, visible: false}
              console.log("AsciinemaPlayer.created "+i.toString(),state.player[i]) // for some reason, without this it does not work
            }
          }
        }
        //console.log(window.location.hash)
        if (window.location.hash.length > 1) {
          state.page = parseInt(window.location.hash.substring(1))-1;
          if (state.page < 0) {
            state.page = 0
          }
          if (state.page >= data.length) {
            state.page = data.length - 1
          }
        }
        if (slides.value != null){
          if (slides.value[state.page].background == ""){
            document.body.style.backgroundImage = 'none'
          }else{
              document.body.style.backgroundImage = 'url('+ slides.value[state.page].background+')'
          }
          document.body.style.fontSize = slides.value[state.page].font_size;
        }
        //console.log(data);
      })
      .catch(error => {
        // Handle any errors here
        console.error('Error:', error);
      });

    let presenterMode = window.location.pathname == '/notes/'
    let timer = ref(`00:00`)
    let counter = 0;
    if (presenterMode) {
      setInterval(function() {
          let minutes = Math.floor(counter / 60);
          let seconds = counter % 60;
          console.log((minutes < 10 ? "0" : "") + minutes + ":" + (seconds < 10 ? "0" : "") + seconds);
          counter++;
          timer.value = (minutes < 10 ? "0" : "") + minutes + ":" + (seconds < 10 ? "0" : "") + seconds
      }, 1000);
    }

    let showMenu = ref(false)

    return {
      slides, md, state, presenterMode, timer, showMenu
    }
  }
});
</script>
