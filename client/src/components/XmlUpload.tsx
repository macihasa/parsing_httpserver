import { useState } from 'react';

const XmlUpload = () => {
  const [selectedFolder, setSelectedFolder] = useState<FileList | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState<string>('');
  const [serverProgress, setServerProgress] = useState<string>('');
  const [file, setFile] = useState<File | null>(null);

  function handleFolderSelection(event: React.ChangeEvent<HTMLInputElement>) {
    const files = event.target.files;
    setSelectedFolder(files);
    console.log(files);
  }

  function handleXmlUpload() {
    if (!selectedFolder || selectedFolder.length === 0) {
      setProgress('No files selected');
      return;
    }

    setIsUploading(true);

    const socket = new WebSocket('ws://localhost:5000/dcecstmsmsg');

    socket.onopen = () => {
      // Send file to server
      for (let i = 0; i < selectedFolder.length; i++) {
        const reader = new FileReader();
        let arrayBuffer: ArrayBuffer;
        reader.onload = () => {
          arrayBuffer = reader.result as ArrayBuffer;
          socket.send(arrayBuffer);
          if (i % 500 == 0) {
            setProgress(`Uploading file ${i} of ${selectedFolder.length}`);
            console.log(`Uploading file ${i} of ${selectedFolder.length}`);
          }
          if (i === selectedFolder.length - 1) {
            setIsUploading(false);
            setProgress('Finished');
            socket.send('Finished');
          }
        };
        reader.readAsArrayBuffer(selectedFolder[i]);
      }
    };

    socket.onmessage = (event) => {
      setServerProgress(event.data);
    };

    socket.onclose = () => {
      console.log('Connection closed: ');
      setServerProgress('Connection closed, requesting file');
      requestFile();
    };
  }

  const requestFile = async () => {
    console.log('Requesting file');

    const response = await fetch('http://localhost:5000/dcecstmsmsg/getfile', {
      method: 'GET',
    });

    const blob = await response.blob();

    // Log content of file
    setFile(new File([blob], 'test.csv'));
    console.log(file);
  };

  return (
    <div className="UploadSegment">
      <div className="UploadButtons">
        <div>
          <input
            type="file"
            multiple={true}
            // @ts-ignore
            webkitdirectory=""
            onChange={handleFolderSelection}
          />
          <button onClick={handleXmlUpload}>
            {isUploading ? 'Uploading...' : 'Submit'}
          </button>
        </div>
        {file && (
          <a href={URL.createObjectURL(file)} download={file.name}>
            Download file
          </a>
        )}
      </div>
      <div className="Logging">
        <h3>{progress}</h3>
        <h3>{serverProgress}</h3>
      </div>
    </div>
  );
};

export default XmlUpload;
