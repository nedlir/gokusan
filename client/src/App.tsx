function App() {
  const callUpload = async () => {
    const res = await fetch("http://localhost:8000/upload");
    const text = await res.text();
    alert(text);
  };

  const callDownload = async () => {
    const res = await fetch("http://localhost:8000/download");
    const text = await res.text();
    alert(text);
  };

  return (
    <div>
      <button onClick={callUpload}>Upload</button>
      <button onClick={callDownload}>Download</button>
    </div>
  );
}

export default App;
