import React from 'react'
import { Invoice } from './data/types'
import InvoicePage from './Pages/InvoicePage'

function App() {
  // const savedInvoice = window.localStorage.getItem('invoiceData')
  let data = undefined


  return (
    <div className="mx-auto w-[700px] mt-8 mb-12">
      <h1 className="center fs-30">Invoice Generator</h1>
      <InvoicePage data={data}  />
    </div>
  )
}

export default App
