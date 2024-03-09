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
      <div class="player-wrap" :id="'ascinema-wrap-' + index">
        <div
          class="player"
          :class="{'zeroSize':!AsciinemaPlayerVisible(index)}"
          :id="'ascinema-player-' + index">
        </div>
      </div>
      <div
        class="terminal"
        v-html="state.terminal[index]"
        v-if="state.terminal[index] != ''">
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
          v-if="state.terminal[index] != ''"
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
      if (this.slides == null) {
        return
      }

      let activeElement = document.activeElement;
      if (activeElement == null) {
        return;
      }
      if (activeElement.classList.contains('code')) {
          return;
      }
      //console.log(activeElement)

      const processData = (data: string) => {
        const playerElement = document.getElementById('ascinema-player-'+this.state.page.toString());
          //const playerWrap = document.getElementById('ascinema-wrap-'+this.state.page.toString());
          if(playerElement!= null) {
            playerElement.innerHTML = '';
          }
          if (playerElement) {
            let  mov = {data: data}
            let jsonStrStart = data.indexOf('{');
            let jsonStrEnd = data.indexOf('}') + 1 + 1;
            let jsonStr = data.slice(jsonStrStart, jsonStrEnd);
            let jsonObject = JSON.parse(jsonStr);
            let width = jsonObject.width;
            let height = jsonObject.height;
            console.log('Width: ' + width + ', Height: ' + height);
            let heightCalc = 30 + (height-3)*3 ;
            if (heightCalc > 75) {
              heightCalc = 75
            }
            playerElement.style.height = heightCalc.toString() + 'vh';
            //playerElement.style.width = '96vw';

            let player = AsciinemaPlayer.create(
              mov,
              playerElement,
              {
                //poster: "data:text/plain,I'm regular \x1b[1;32mI'm bold green\x1b[3BI'm 3 lines down",
                fit: "both",
                poster: 'npt:0:0',
                width: width,
                height: height,
                theme: "solarized-light",
                controls: false,
                //terminalFontSize: ""
              })
            player.play();
            //absolute

            //playerElement.style.width = 'unset';
            setTimeout(function() {
              //playerElement.style.width = 'unset';
              //playerElement.style.height = 'unset';
            playerElement.style.position = 'absolute';
            }, 1); // delay in milliseconds

            console.log(player)
            this.state.player[this.state.page] = {player: player, visible: true}
            console.log("AsciinemaPlayer.created "+this.state.page.toString(),this.state.player[this.state.page]) // for some reason, without this it does not work
          }
          return data;
      }

      if (this.slides[this.state.page].cast) {
        const baseUrl = import.meta.env.VITE_BASE_URL
        const canEdit = this.slides[this.state.page].can_edit
        if (canEdit){
          const slideElement = document.getElementById('slide-'+this.state.page);
          let codeText: string[] = [];
          if (slideElement) {
            let codeElements = slideElement.querySelectorAll('pre code');
            codeText = Array.from(codeElements).map(codeElement => (codeElement as HTMLElement).innerText);
          }
          //console.log(codeText)
          let body = JSON.stringify({slide: this.state.page, code: codeText})
          console.log(body)
          fetch(baseUrl + "/cast?slide=" + this.state.page, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: body
          })
          .then(response => response.text())
          .then(data => {
            return processData(data)
          })
          return
        }
        fetch(baseUrl+"/cast?slide="+this.state.page)
        .then(response => response.text())
        .then(data => {
            return processData(data)
        })
        .catch(error => {
          // Handle any errors here
          console.error('Error:', error);
        });
        // do not execute terminal run
        return
      }

      // check if we have a asciinema file
      let pl = this.state.player[this.state.page]
      if (pl.player!= null) {
        pl.player.play();
        pl.visible = true;
        // do not execute terminal run
        return
      }
      if (!this.slides[this.state.page].terminal) {
        return
      }
      this.state.terminal[this.state.page] = "Running..."
      const baseUrl = import.meta.env.VITE_BASE_URL
      fetch(baseUrl+"/exec?slide="+this.state.page)
      .then(response => response.text())
      .then(data => {
        data = data.replace(/\n/g, "<br>");
        data = data.replace(/ /g, "&nbsp;");
        this.state.terminal[this.state.page] = data;
        return data;
      })
      /*.then(data => {
        console.log(data);
      })*/
      .catch(error => {
        // Handle any errors here
        console.error('Error:', error);
      });
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
    AsciinemaPlayerVisible: function (index: number) {
      if (index==null || this.slides == null) {
        return false
      }
      //slide.asciinema=='' state.player[state.page].visible
      let state = this.state
      if (state.player[index] == null) {
          return false
      }
      let pl = state.player[index]
      if (pl.visible) {
        return true
      }
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
        // check if we have a asciinema file
        let pl = this.state.player[this.state.page]
        if (pl.player!= null) {
          pl.player.pause()
          pl.visible = false
          return
        }
      }
      if (keyCode == ' ') {
        // check if we have a asciinema file
        let pl = this.state.player[this.state.page]
        if (pl.player!= null) {
          pl.player.pause()
          return
        }
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
          let codeElements = codeSelector.querySelectorAll('pre code')
          codeElements.forEach((codeElement) => {
            (codeElement as HTMLElement).contentEditable = "true";
            (codeElement as HTMLElement).spellcheck = false;
          })
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
      asciinema: any;
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
