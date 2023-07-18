import './App.css';
import ExcelUpload from './components/ExcelUpload';
import XmlUpload from './components/XmlUpload';

const App: React.FC = () => {
  return (
    <div className="App">
      <h2>DCC XML files - Websocket</h2>
      <XmlUpload />
      <br />
      <h2>Excel files - HTTP Post</h2>
      <ExcelUpload />
      <br />
      <br />
      <p>
        Toolkit to combine different document types together into a consolidated
        csv file. <br />
        For excel files: enter the name of the sheets you'd like to combine.
        Alternatively enter "1" for the first sheet of each file <br />
      </p>
    </div>
  );
};

export default App;
