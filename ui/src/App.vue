<script setup lang="ts">
import Slide from './components/Slide.vue'
import { ref, onMounted, onUnmounted } from 'vue';

let syncMessage = ref(Object);

let socket: WebSocket = null as unknown as WebSocket;

function startWebsocket() {
    const baseUrlStart = import.meta.env.VITE_WS_URL.replace(/^https?:\/\//, '');
    const baseUrl = baseUrlStart.replace(/^(\/|\\)/, '');
    socket = new WebSocket('ws://'+baseUrl+"/ws");

    // Connection opened
    socket.addEventListener('open', () => {
      //console.log(codeText)
      //let body = JSON.stringify({slide: this.state.page, code: codeText})
      //console.log(body)
      //socket.send(body);
    });

    // Listen for messages
    socket.addEventListener('message', (event) => {
        console.log('Message from server: ', event.data);
        const data = JSON.parse(event.data)
        if (data.Reload) {
          location.reload()
        } else {
          if (data.Slide > -999){
            syncMessage.value = data
          }
        }
    });
    
    socket.onclose = () => {
      console.log('Socket is closed');
      socket.close(); 
      setTimeout(startWebsocket, 5000)    
    };
  }

onMounted(() => {
  startWebsocket();  
});


onUnmounted(() => {
});

const updateData = (data : any) => {
    console.log("updateData send", data)
    
    if (socket === null) {
      console.log("socket is null, skipping send")
      return
    }

    // if (firstMessage) {
    //   console.log("first message, skipping send")
    //   firstMessage = false
    //   return
    // }
    socket.send(JSON.stringify(data))
};

</script>

<template>
  <Slide :syncMessage="syncMessage" @updateData="updateData"/>
</template>
