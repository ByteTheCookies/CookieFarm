import { useState, useEffect } from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { arta } from 'react-syntax-highlighter/dist/esm/styles/hljs';



export function CodeTypewriter(props: Readonly<{ code: string }>) {
  const fullCode = props.code;
  const [displayed, setDisplayed] = useState('');
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (index >= fullCode.length) return;
    const timer = setTimeout(() => {
      setDisplayed(fullCode.slice(0, index + 1));
      setIndex(i => i + 1);
    }, 40); // ms per character
    return () => clearTimeout(timer);
  }, [index]);

  return (
    <SyntaxHighlighter language="python" showLineNumbers={true} style={arta} customStyle={{ background: 'transparent', padding: '1rem', borderRadius: '0.5rem' }}>
      {displayed}
    </SyntaxHighlighter>
  );
}
