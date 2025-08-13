import { useState } from 'react';
import {
  User,
  Folder,
  Coffee,
  Star,
  AlertCircle,
  CheckCircle,
  Loader2
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useProfile } from '@/contexts/ProfileContext';
import { useToast, toast } from '@/components/ui/toast';
import { CreateProfileRequest } from '@/types';

interface OnboardingProfileStepProps {
  onProfileCreated: (profile: any) => void;
  isProcessing: boolean;
  setIsProcessing: (processing: boolean) => void;
}

export function OnboardingProfileStep({
  onProfileCreated,
  isProcessing,
  setIsProcessing
}: OnboardingProfileStepProps) {
  const { createProfile } = useProfile();
  const { addToast } = useToast();

  const [formData, setFormData] = useState<CreateProfileRequest>({
    name: '',
    description: '',
    services: [],
    envVars: {},
    projectsDir: '',
    javaHomeOverride: '',
    isDefault: true, // Make first profile default
    isActive: true, // Set as active during onboarding
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Profile name is required';
    } else if (formData.name.length < 2) {
      newErrors.name = 'Profile name must be at least 2 characters';
    }

    if (!formData.projectsDir.trim()) {
      newErrors.projectsDir = 'Projects directory is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validateForm()) return;

    setIsProcessing(true);
    try {
      const profile = await createProfile(formData);
      
      addToast(
        toast.success(
          'Profile Created!',
          `"${formData.name}" profile has been created successfully.`
        )
      );

      onProfileCreated(profile);
    } catch (error) {
      console.error('Failed to create profile:', error);
      addToast(
        toast.error(
          'Failed to create profile',
          error instanceof Error ? error.message : 'An unexpected error occurred'
        )
      );
    } finally {
      setIsProcessing(false);
    }
  };

  const handleInputChange = (field: keyof CreateProfileRequest, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error for this field
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center">
        <div className="mx-auto w-16 h-16 bg-blue-100 dark:bg-blue-900/20 rounded-full flex items-center justify-center mb-4">
          <User className="w-8 h-8 text-blue-600 dark:text-blue-400" />
        </div>
        <h3 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
          Create Your First Profile
        </h3>
        <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
          Profiles help you organize and manage different environments or projects. 
          Let's start by creating your main workspace profile.
        </p>
      </div>

      {/* Form */}
      <Card className="max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Star className="w-5 h-5 text-yellow-500" />
            Profile Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Profile Name */}
          <div>
            <Label htmlFor="profileName" className="flex items-center gap-2">
              <User className="w-4 h-4" />
              Profile Name *
            </Label>
            <Input
              id="profileName"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              placeholder="e.g., Development, Staging, Main Workspace"
              className={errors.name ? 'border-red-500' : ''}
            />
            {errors.name && (
              <div className="flex items-center gap-1 mt-1 text-sm text-red-600">
                <AlertCircle className="w-4 h-4" />
                {errors.name}
              </div>
            )}
          </div>

          {/* Description */}
          <div>
            <Label htmlFor="profileDescription">
              Description (Optional)
            </Label>
            <Textarea
              id="profileDescription"
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              placeholder="Brief description of this profile's purpose..."
              rows={3}
            />
          </div>

          {/* Projects Directory */}
          <div>
            <Label htmlFor="projectsDir" className="flex items-center gap-2">
              <Folder className="w-4 h-4" />
              Projects Directory *
            </Label>
            <Input
              id="projectsDir"
              value={formData.projectsDir}
              onChange={(e) => handleInputChange('projectsDir', e.target.value)}
              placeholder="/path/to/your/microservices/projects"
              className={errors.projectsDir ? 'border-red-500' : ''}
            />
            {errors.projectsDir && (
              <div className="flex items-center gap-1 mt-1 text-sm text-red-600">
                <AlertCircle className="w-4 h-4" />
                {errors.projectsDir}
              </div>
            )}
            <p className="text-sm text-gray-500 mt-1">
              The root directory where your microservice projects are located
            </p>
          </div>

          {/* Java Home Override */}
          <div>
            <Label htmlFor="javaHome" className="flex items-center gap-2">
              <Coffee className="w-4 h-4" />
              Java Home Override (Optional)
            </Label>
            <Input
              id="javaHome"
              value={formData.javaHomeOverride}
              onChange={(e) => handleInputChange('javaHomeOverride', e.target.value)}
              placeholder="/path/to/java/home (optional)"
            />
            <p className="text-sm text-gray-500 mt-1">
              Override the default Java installation for this profile
            </p>
          </div>

          {/* Default Profile */}
          <div className="flex items-center space-x-2 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <Checkbox
              id="isDefault"
              checked={formData.isDefault}
              onCheckedChange={(checked) => handleInputChange('isDefault', checked === true)}
            />
            <Label htmlFor="isDefault" className="flex items-center gap-2">
              <Star className="w-4 h-4 text-yellow-500" />
              Set as default profile
            </Label>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              (Recommended for your first profile)
            </p>
          </div>

          {/* Submit Button */}
          <div className="pt-4">
            <Button
              onClick={handleSubmit}
              disabled={isProcessing}
              className="w-full"
              size="lg"
            >
              {isProcessing ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Creating Profile...
                </>
              ) : (
                <>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  Create Profile
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Tips */}
      <Card className="max-w-2xl mx-auto bg-gray-50 dark:bg-gray-800">
        <CardContent className="p-4">
          <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
            💡 Profile Tips
          </h4>
          <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
            <li>• Profiles help you organize services by environment (dev, staging, prod)</li>
            <li>• Each profile can have its own projects directory and Java settings</li>
            <li>• You can create multiple profiles and switch between them easily</li>
            <li>• The default profile is automatically selected when you log in</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

export default OnboardingProfileStep;