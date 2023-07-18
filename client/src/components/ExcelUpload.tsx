import { useState } from 'react';

const ExcelUpload = () => {
  const [selectedFolder, setSelectedFolder] = useState<FileList | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState<string>('');
  const [serverProgress, setServerProgress] = useState<string>('');
  const [sheetName, setSheetName] = useState('');
  const [file, setFile] = useState<File | null>(null);

  function handleFolderSelection(event: React.ChangeEvent<HTMLInputElement>) {
    const files = event.target.files;
    setSelectedFolder(files);
    console.log(files);
  }

  async function handleExcelUpload() {
    let formdata = new FormData();

    if (selectedFolder == null) {
      setProgress('Please select files');
      return;
    }

    for (let i = 0; i < selectedFolder.length; i++) {
      formdata.append('files', selectedFolder[i], selectedFolder[i].name);
    }

    console.log('http://localhost:5000/excel/' + sheetName);

    const res = await fetch('http://localhost:5000/excel/' + sheetName, {
      method: 'POST',
      body: formdata,
    });

    const data = await res.blob();
    setFile(new File([data], 'test.csv'));
  }

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSheetName(event.target.value);
  };

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
          <button onClick={handleExcelUpload}>
            {isUploading ? 'Uploading...' : 'Submit'}
          </button>
          <br />
          <input
            type="text"
            value={sheetName}
            onChange={handleInputChange}
            placeholder="Enter sheet name:"
            style={{ width: '100%' }}
            required
          />
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

export default ExcelUpload;
