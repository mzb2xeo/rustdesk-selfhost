export const connectByClient = (id) => {
  //Open the url protocol without opening a new window, the format is rustdesk://<id>
  // window.open(`rustdesk://${row.id}`)
  let a = document.createElement('a')
  a.href = `rustdesk://${id}`
  a.target = '_self'
  a.click()

}
