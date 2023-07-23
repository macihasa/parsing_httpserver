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

  function handleExcelUpload() {
    requestFile().catch((err) => {
      console.log(err);
    });
  }

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSheetName(event.target.value);
  };

  const requestFile = async () => {
    const formdata = new FormData();

    if (selectedFolder == null) {
      setProgress('Please select files');
      return;
    }

    for (let i = 0; i < selectedFolder.length; i++) {
      formdata.append('files', selectedFolder[i], selectedFolder[i].name);
    }

    console.log('http://macihasa.com:5000/excel/' + sheetName);

    setIsUploading(true);
    setServerProgress('Uploading to server...');
    const res = await fetch('http://macihasa.com:5000/excel/' + sheetName, {
      method: 'POST',
      body: formdata,
    });

    const data = await res.blob();
    setFile(new File([data], 'test.csv'));
    setServerProgress('Server finished processing: ');
  };

  return (
    <div className="UploadSegment">
      <div className="UploadButtons">
        <div>
          <input
            type="file"
            multiple={true}
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
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
