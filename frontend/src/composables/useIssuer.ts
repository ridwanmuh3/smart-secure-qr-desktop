import { ref, computed } from 'vue'
import type { QRGenerationResult, IssuerConfig, KeyPairInfo } from '../types'
import { SelectDocument, HashDocumentFile, GetDocumentInfo, GenerateSecureQR, SaveQRImage, SaveSignedDocument, ListKeyPairs } from '../../wailsjs/go/main/App'

export function useIssuer() {
  const filePath = ref('')
  const fileName = ref('')
  const fileSize = ref(0)
  const documentHash = ref('')
  const validFrom = ref('')
  const validUntil = ref('')
  const selectedKeyId = ref('')
  const metadata = ref('')
  const issuerID = ref('')
  const qrPosition = ref('bottom-right')
  const qrPage = ref(0) // 0 = last page
  const qrSize = ref(30) // mm
  const isGenerating = ref(false)
  const result = ref<QRGenerationResult | null>(null)
  const error = ref('')
  const keys = ref<KeyPairInfo[]>([])

  function setTimePreset(preset: string) {
    const now = new Date()
    validFrom.value = formatDatetimeLocal(now)

    const end = new Date(now)
    switch (preset) {
      case 'hour': end.setHours(end.getHours() + 1); break
      case 'day': end.setDate(end.getDate() + 1); break
      case 'week': end.setDate(end.getDate() + 7); break
      case 'month': end.setMonth(end.getMonth() + 1); break
    }
    validUntil.value = formatDatetimeLocal(end)
  }

  function formatDatetimeLocal(d: Date): string {
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }

  async function selectDocument() {
    error.value = ''
    try {
      const path = await SelectDocument()
      if (!path) return
      filePath.value = path
      const info = await GetDocumentInfo(path)
      fileName.value = info.name as string
      fileSize.value = info.size as number
      documentHash.value = await HashDocumentFile(path)
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function loadKeys() {
    try {
      const list = await ListKeyPairs()
      keys.value = list || []
      const defaultKey = keys.value.find(k => k.is_default)
      if (defaultKey && !selectedKeyId.value) {
        selectedKeyId.value = defaultKey.id
      }
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function generateQR() {
    error.value = ''
    if (!filePath.value) { error.value = 'Pilih dokumen terlebih dahulu'; return }
    if (!selectedKeyId.value) { error.value = 'Pilih kunci terlebih dahulu'; return }
    if (!validFrom.value || !validUntil.value) { error.value = 'Tentukan jendela waktu'; return }

    isGenerating.value = true
    try {
      const config: IssuerConfig = {
        file_path: filePath.value,
        key_pair_id: selectedKeyId.value,
        valid_from: validFrom.value,
        valid_until: validUntil.value,
        metadata: metadata.value,
        issuer_id: issuerID.value,
        qr_position: qrPosition.value,
        qr_page: qrPage.value,
        qr_size: qrSize.value,
      }
      result.value = await GenerateSecureQR(config)
    } catch (e: any) {
      error.value = e?.message || String(e)
    } finally {
      isGenerating.value = false
    }
  }

  const isPDF = computed(() => filePath.value.toLowerCase().endsWith('.pdf'))

  async function saveQR() {
    if (!result.value?.qr_code_base64) return
    try {
      await SaveQRImage(result.value.qr_code_base64)
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  async function saveSignedDoc() {
    if (!result.value?.signed_file_path) return
    try {
      await SaveSignedDocument(result.value.signed_file_path)
    } catch (e: any) {
      error.value = e?.message || String(e)
    }
  }

  function reset() {
    filePath.value = ''
    fileName.value = ''
    fileSize.value = 0
    documentHash.value = ''
    metadata.value = ''
    result.value = null
    error.value = ''
  }

  // Set default time window
  setTimePreset('day')

  return {
    filePath, fileName, fileSize, documentHash,
    validFrom, validUntil, selectedKeyId, metadata, issuerID,
    qrPosition, qrPage, qrSize, isPDF,
    isGenerating, result, error, keys,
    selectDocument, loadKeys, generateQR, saveQR, saveSignedDoc, reset, setTimePreset,
  }
}
