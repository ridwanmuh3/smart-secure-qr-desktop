<script lang="ts" setup>
import { onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useIssuer } from '../composables/useIssuer'
import FileUploader from '../components/issuer/FileUploader.vue'
import TimeWindowConfig from '../components/issuer/TimeWindowConfig.vue'
import KeySelector from '../components/issuer/KeySelector.vue'
import QRPositionConfig from '../components/issuer/QRPositionConfig.vue'
import QRCodeDisplay from '../components/issuer/QRCodeDisplay.vue'

const {
  filePath, fileName, fileSize, documentHash,
  validFrom, validUntil, selectedKeyId, metadata, issuerID,
  qrPosition, qrPage, qrSize, isPDF,
  isGenerating, result, error, keys,
  selectDocument, loadKeys, generateQR, saveQR, saveSignedDoc, reset, setTimePreset,
} = useIssuer()

onMounted(() => loadKeys())
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="mb-6">
      <h2 class="text-2xl font-bold text-gray-800">Buat QR Code Aman</h2>
      <p class="text-gray-500 text-sm mt-1">Tandatangani dokumen dan buat QR code terenkripsi</p>
    </div>

    <!-- Error alert -->
    <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm flex items-center gap-2">
      <Icon icon="lucide:circle-alert" class="w-5 h-5 shrink-0" />
      <span>{{ error }}</span>
    </div>

    <!-- Result view -->
    <template v-if="result?.success">
      <QRCodeDisplay
        :qr-base64="result.qr_code_base64"
        :secure-id="result.secure_id"
        :document-hash="result.document_hash || documentHash"
        :file-name="fileName"
        :valid-from="validFrom"
        :valid-until="validUntil"
        :signed-file-path="result.signed_file_path"
        :is-pdf="result.is_pdf"
        @save="saveQR"
        @save-signed="saveSignedDoc"
        @reset="reset"
      />
    </template>

    <!-- Form view -->
    <template v-else>
      <div class="space-y-5">
        <!-- Step 1: File -->
        <FileUploader
          :file-name="fileName"
          :file-size="fileSize"
          :document-hash="documentHash"
          @select="selectDocument"
        />

        <!-- Step 2: Time Window -->
        <TimeWindowConfig
          v-model:valid-from="validFrom"
          v-model:valid-until="validUntil"
          @preset="setTimePreset"
        />

        <!-- Step 3: Key Selection -->
        <KeySelector
          v-model="selectedKeyId"
          :keys="keys"
          @refresh="loadKeys"
        />

        <!-- Step 4: QR Position on PDF -->
        <QRPositionConfig
          v-if="filePath"
          v-model:position="qrPosition"
          v-model:page="qrPage"
          v-model:size="qrSize"
          :is-pdf="isPDF"
        />

        <!-- Step 5: Metadata -->
        <div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <h3 class="text-sm font-semibold text-gray-700 mb-3">Informasi Tambahan</h3>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs text-gray-500 mb-1">ID Penerbit</label>
              <input v-model="issuerID" type="text" placeholder="Nama atau ID penerbit"
                class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700" />
            </div>
            <div>
              <label class="block text-xs text-gray-500 mb-1">Metadata (opsional)</label>
              <input v-model="metadata" type="text" placeholder="Catatan tambahan"
                class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700" />
            </div>
          </div>
        </div>

        <!-- Generate Button -->
        <button
          @click="generateQR"
          :disabled="isGenerating || !filePath || !selectedKeyId"
          class="w-full py-3 bg-emerald-600 hover:bg-emerald-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white font-semibold rounded-xl shadow-md transition-all duration-200 flex items-center justify-center gap-2"
        >
          <Icon v-if="isGenerating" icon="lucide:loader-circle" class="w-5 h-5" />
          <Icon v-else icon="lucide:plus" class="w-5 h-5" />
          {{ isGenerating ? 'Memproses...' : 'Buat QR Code & PDF' }}
        </button>
      </div>
    </template>
  </div>
</template>
