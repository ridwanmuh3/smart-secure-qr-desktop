import { ref } from 'vue'
import type { VerificationResult } from '../types'
import { SelectQRImage, VerifyQRFromImage, VerifyQRFromData, VerifyDocument } from '../../wailsjs/go/main/App'

export function useVerifier() {
  const isVerifying = ref(false)
  const verificationResult = ref<VerificationResult | null>(null)
  const error = ref('')
  const activeTab = ref<'document' | 'image' | 'camera'>('document')

  async function verifyFromDocument() {
    error.value = ''
    isVerifying.value = true
    try {
      verificationResult.value = await VerifyDocument() as unknown as VerificationResult
    } catch (e: any) {
      error.value = e?.message || String(e)
    } finally {
      isVerifying.value = false
    }
  }

  async function verifyFromImage() {
    error.value = ''
    isVerifying.value = true
    try {
      const path = await SelectQRImage()
      if (!path) { isVerifying.value = false; return }
      verificationResult.value = await VerifyQRFromImage(path) as unknown as VerificationResult
    } catch (e: any) {
      error.value = e?.message || String(e)
    } finally {
      isVerifying.value = false
    }
  }

  async function verifyFromData(qrData: string) {
    error.value = ''
    isVerifying.value = true
    try {
      verificationResult.value = await VerifyQRFromData(qrData) as unknown as VerificationResult
    } catch (e: any) {
      error.value = e?.message || String(e)
    } finally {
      isVerifying.value = false
    }
  }

  function reset() {
    verificationResult.value = null
    error.value = ''
  }

  return {
    isVerifying, verificationResult, error, activeTab,
    verifyFromDocument, verifyFromImage, verifyFromData, reset,
  }
}
