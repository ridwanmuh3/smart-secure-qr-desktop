<script lang="ts" setup>
  import { Icon } from '@iconify/vue'
  import { useVerifier } from '../composables/useVerifier'
  import QRScanner from '../components/verifier/QRScanner.vue'
  import VerificationResultCard from '../components/verifier/VerificationResult.vue'

  const {
    isVerifying, verificationResult, error, activeTab,
    verifyFromDocument, verifyFromImage, verifyFromData, reset,
  } = useVerifier()
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="mb-6">
      <h2 class="text-2xl font-bold text-gray-800">Verifikasi Dokumen</h2>
      <p class="text-gray-500 text-sm mt-1">Unggah dokumen PDF atau gambar QR code untuk memverifikasi keaslian</p>
    </div>

    <!-- Error alert -->
    <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
      {{ error }}
    </div>

    <!-- Result display -->
    <template v-if="verificationResult">
      <VerificationResultCard :result="verificationResult" @reset="reset" />
    </template>

    <!-- Scanner/Upload tabs -->
    <template v-else>
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <!-- Tab header -->
        <div class="flex border-b border-gray-200">
          <button @click="activeTab = 'document'"
            class="flex-1 py-3 px-4 text-sm font-medium transition-colors cursor-pointer"
            :class="activeTab === 'document' ? 'text-emerald-700 border-b-2 border-emerald-600 bg-emerald-50' : 'text-gray-500 hover:text-gray-700'">
            <span class="flex items-center justify-center gap-2">
              <Icon icon="lucide:file-text" class="w-4 h-4" />
              Unggah Dokumen
            </span>
          </button>
          <button @click="activeTab = 'image'"
            class="flex-1 py-3 px-4 text-sm font-medium transition-colors cursor-pointer"
            :class="activeTab === 'image' ? 'text-emerald-700 border-b-2 border-emerald-600 bg-emerald-50' : 'text-gray-500 hover:text-gray-700'">
            <span class="flex items-center justify-center gap-2">
              <Icon icon="lucide:image" class="w-4 h-4" />
              Unggah Gambar QR
            </span>
          </button>
          <button @click="activeTab = 'camera'"
            class="flex-1 py-3 px-4 text-sm font-medium transition-colors cursor-pointer"
            :class="activeTab === 'camera' ? 'text-emerald-700 border-b-2 border-emerald-600 bg-emerald-50' : 'text-gray-500 hover:text-gray-700'">
            <span class="flex items-center justify-center gap-2">
              <Icon icon="lucide:camera" class="w-4 h-4" />
              Pindai Kamera
            </span>
          </button>
        </div>

        <!-- Tab content -->
        <div class="p-6">
          <!-- Upload document tab -->
          <div v-if="activeTab === 'document'" class="text-center">
            <div
              class="border-2 border-dashed border-gray-300 rounded-xl p-10 hover:border-emerald-400 transition-colors">
              <Icon icon="lucide:file-text" class="w-12 h-12 text-gray-400 mx-auto mb-3" />
              <p class="text-gray-600 mb-2">Pilih dokumen PDF yang mengandung QR code</p>
              <p class="text-gray-400 text-xs mb-4">Sistem akan mendeteksi dan memverifikasi QR code secara otomatis</p>
              <button @click="verifyFromDocument" :disabled="isVerifying"
                class="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-700 disabled:bg-gray-300 text-white font-medium rounded-lg transition-colors inline-flex items-center gap-2 cursor-pointer">
                <Icon v-if="isVerifying" icon="lucide:loader-circle" class="w-4 h-4" />
                {{ isVerifying ? 'Memverifikasi...' : 'Pilih Dokumen PDF' }}
              </button>
            </div>
          </div>

          <!-- Upload image tab -->
          <div v-if="activeTab === 'image'" class="text-center">
            <div
              class="border-2 border-dashed border-gray-300 rounded-xl p-10 hover:border-emerald-400 transition-colors">
              <Icon icon="lucide:image" class="w-12 h-12 text-gray-400 mx-auto mb-3" />
              <p class="text-gray-600 mb-4">Pilih gambar QR code untuk diverifikasi</p>
              <button @click="verifyFromImage" :disabled="isVerifying"
                class="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-700 disabled:bg-gray-300 text-white font-medium rounded-lg transition-colors inline-flex items-center gap-2 cursor-pointer">
                <Icon v-if="isVerifying" icon="lucide:loader-circle" class="w-4 h-4" />
                {{ isVerifying ? 'Memverifikasi...' : 'Pilih Gambar QR' }}
              </button>
            </div>
          </div>

          <!-- Camera tab -->
          <div v-if="activeTab === 'camera'">
            <QRScanner @decoded="verifyFromData" />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
