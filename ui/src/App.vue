<script setup lang="ts">
import Slide from './components/Slide.vue'
import { ref, onMounted, onUnmounted, Ref } from 'vue';

const eventSource: Ref<EventSource | null> = ref(null);
let syncMessage = ref(Object);

onMounted(() => {
  const baseUrl = import.meta.env.VITE_BASE_URL
  eventSource.value = new EventSource(baseUrl+ '/sync');
  eventSource.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    syncMessage.value = data
    console.log("event received from server "+syncMessage.value)
    if (data.Reload) {
      location.reload()
    }
  };
});

onUnmounted(() => {
  if (eventSource.value) {
    eventSource.value.close();
  }
});

</script>

<template>
  <Slide :syncMessage="syncMessage" />
</template>
