import InvoicePage from './Pages/InvoicePage'

function App() {
  let data = undefined

  return (
    <div className="mx-auto w-[700px] mt-8 mb-12">
      <h1 className="center fs-30">Invoice Generator</h1>
      <InvoicePage data={data}  />
    </div>
  )
}

export default App
