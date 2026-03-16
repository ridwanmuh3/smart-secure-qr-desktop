<script lang="ts" setup>
  import { onMounted, onBeforeUnmount, ref } from 'vue'
  import { Html5Qrcode } from 'html5-qrcode'
  import { Icon } from '@iconify/vue'

  const emit = defineEmits<{
    decoded: [data: string]
  }>()

  const scannerRef = ref<Html5Qrcode | null>(null)
  const scannerError = ref('')
  const isStarting = ref(true)

  onMounted(async () => {
    try {
      const scanner = new Html5Qrcode('qr-reader')
      scannerRef.value = scanner

      await scanner.start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: { width: 250, height: 250 } },
        (decodedText) => {
          scanner.stop().catch(() => { })
          emit('decoded', decodedText)
        },
        () => { } // ignore errors during scanning
      )
      isStarting.value = false
    } catch (e: any) {
      isStarting.value = false
      scannerError.value = e?.message || 'Gagal mengakses kamera. Pastikan izin kamera diberikan.'
    }
  })

  onBeforeUnmount(() => {
    scannerRef.value?.stop().catch(() => { })
  })
</script>

<template>
  <div>
    <div v-if="scannerError" class="text-center py-8">
      <Icon icon="lucide:video-off" class="w-12 h-12 text-gray-400 mx-auto mb-3" />
      <p class="text-red-500 text-sm">{{ scannerError }}</p>
    </div>

    <div v-if="isStarting && !scannerError" class="text-center py-8">
      <Icon icon="lucide:loader-circle" class="w-8 h-8 text-emerald-500 mx-auto mb-2" />
      <p class="text-gray-500 text-sm">Memulai kamera...</p>
    </div>

    <div id="qr-reader" class="rounded-lg overflow-hidden"></div>
    <p class="text-xs text-gray-400 text-center mt-3">Arahkan kamera ke QR code</p>
  </div>
</template>
