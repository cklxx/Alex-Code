export interface GenerationResult {
  code: string;
  language: string;
  explanation?: string;
}

export class CodeGenerator {
  private templates: Map<string, Map<string, string>>;

  constructor() {
    this.templates = new Map();
    this.initializeTemplates();
  }

  async generate(specification: string, language: string = 'typescript'): Promise<GenerationResult> {
    const normalizedLang = language.toLowerCase();
    
    if (this.isApiRequest(specification)) {
      return this.generateApiCode(specification, normalizedLang);
    }
    
    if (this.isComponentRequest(specification)) {
      return this.generateComponent(specification, normalizedLang);
    }
    
    if (this.isUtilityRequest(specification)) {
      return this.generateUtility(specification, normalizedLang);
    }
    
    return this.generateGenericCode(specification, normalizedLang);
  }

  private initializeTemplates(): void {
    const tsTemplates = new Map([
      ['api', `import express from 'express';

const router = express.Router();

router.get('/', (req, res) => {
  res.json({ message: 'Hello from API' });
});

export default router;`],
      
      ['component', `import React from 'react';

interface Props {
  title: string;
}

export const Component: React.FC<Props> = ({ title }) => {
  return (
    <div>
      <h1>{title}</h1>
    </div>
  );
};`],
      
      ['utility', `export const utility = (input: any): any => {
  // TODO: Implement utility function
  return input;
};`],
      
      ['class', `export class ClassName {
  private property: string;

  constructor(property: string) {
    this.property = property;
  }

  public method(): string {
    return this.property;
  }
}`]
    ]);

    const pyTemplates = new Map([
      ['api', `from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/')
def hello():
    return jsonify({"message": "Hello from API"})

if __name__ == '__main__':
    app.run(debug=True)`],
      
      ['utility', `def utility_function(input_data):
    """
    Utility function description
    """
    # TODO: Implement utility function
    return input_data`],
      
      ['class', `class ClassName:
    def __init__(self, property_value):
        self.property = property_value
    
    def method(self):
        return self.property`]
    ]);

    this.templates.set('typescript', tsTemplates);
    this.templates.set('javascript', tsTemplates);
    this.templates.set('python', pyTemplates);
  }

  private isApiRequest(spec: string): boolean {
    const apiKeywords = ['api', 'endpoint', 'route', 'server', 'rest'];
    return apiKeywords.some(keyword => spec.toLowerCase().includes(keyword));
  }

  private isComponentRequest(spec: string): boolean {
    const componentKeywords = ['component', 'react', 'ui', 'interface'];
    return componentKeywords.some(keyword => spec.toLowerCase().includes(keyword));
  }

  private isUtilityRequest(spec: string): boolean {
    const utilityKeywords = ['utility', 'helper', 'function', 'util'];
    return utilityKeywords.some(keyword => spec.toLowerCase().includes(keyword));
  }

  private generateApiCode(spec: string, language: string): GenerationResult {
    const template = this.getTemplate(language, 'api');
    
    return {
      code: this.customizeTemplate(template, spec),
      language,
      explanation: 'Generated API endpoint with basic structure'
    };
  }

  private generateComponent(spec: string, language: string): GenerationResult {
    const template = this.getTemplate(language, 'component');
    
    return {
      code: this.customizeTemplate(template, spec),
      language,
      explanation: 'Generated React component with props interface'
    };
  }

  private generateUtility(spec: string, language: string): GenerationResult {
    const template = this.getTemplate(language, 'utility');
    
    return {
      code: this.customizeTemplate(template, spec),
      language,
      explanation: 'Generated utility function with basic structure'
    };
  }

  private generateGenericCode(spec: string, language: string): GenerationResult {
    const template = this.getTemplate(language, 'class');
    
    return {
      code: this.customizeTemplate(template, spec),
      language,
      explanation: 'Generated generic class structure based on specification'
    };
  }

  private getTemplate(language: string, type: string): string {
    const langTemplates = this.templates.get(language);
    if (!langTemplates) {
      return this.templates.get('typescript')?.get(type) || '// Template not found';
    }
    
    return langTemplates.get(type) || langTemplates.get('utility') || '// Template not found';
  }

  private customizeTemplate(template: string, spec: string): string {
    const words = spec.split(' ');
    const capitalizedWords = words.map(word => 
      word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()
    );
    
    const className = capitalizedWords.join('');
    const functionName = words.join('_').toLowerCase();
    
    return template
      .replace(/ClassName/g, className)
      .replace(/utility_function/g, functionName)
      .replace(/utility/g, functionName)
      .replace(/Component/g, className)
      .replace(/TODO: Implement.*$/gm, `// Implementation for: ${spec}`);
  }
}