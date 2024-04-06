import { CSSProperties } from 'react'

export interface ProductLine {
  description: string
  quantity: number
  rate: number
}

export interface Invoice {
  logo: string
  logoWidth: number
  title: string
  companyName: string
  name: string
  companyAddress: string
  companyAddress2: string
  companyCountry: string
  companyGST: string

  billTo: string
  clientName: string
  clientAddress: string
  clientAddress2: string
  clientCountry: string
  clientGST: string

  invoiceTitleLabel: string
  invoiceTitle: string
  invoiceDateLabel: string
  invoiceDate: string
  invoiceDueDateLabel: string
  invoiceDueDate: string

  productLineDescription: string
  productLineQuantity: string
  productLineQuantityRate: string
  productLineQuantityAmount: string

  productLines: ProductLine[]

  subTotalLabel: string
  taxLabel1: string
  taxLabel2: string
  taxPercentage1: string
  taxPercentage2: string

  totalLabel: string
  currency: string

  notesLabel: string
  notes: string
  termLabel: string
  term: string
}

export interface CSSClasses {
  [key: string]: CSSProperties
}
